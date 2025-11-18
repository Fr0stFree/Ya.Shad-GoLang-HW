package ledger_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"gitlab.com/slon/shad-go/ledger"
	"gitlab.com/slon/shad-go/pgfixture"
)

func TestLedger(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) })

	dsn := pgfixture.Start(t)
	ctx := context.Background()

	l0, err := ledger.New(ctx, dsn)
	require.NoError(t, err)
	defer func() { _ = l0.Close() }()

	t.Run("SimpleCommands", func(t *testing.T) {
		testSimpleCommands(t, ctx, l0)
	})

	t.Run("ErroneousCases", func(t *testing.T) {
		testErroneousCases(t, ctx, l0)
	})

	t.Run("Transactions", func(t *testing.T) {
		testTransactions(t, ctx, l0)
	})
}

func testSimpleCommands(t *testing.T, ctx context.Context, l0 ledger.Ledger) {
	checkBalance := func(account ledger.ID, amount ledger.Money) {
		b, err := l0.GetBalance(ctx, account)
		require.NoError(t, err)
		require.Equal(t, amount, b)
	}

	require.NoError(t, l0.CreateAccount(ctx, "a0"))
	checkBalance("a0", 0)

	require.Error(t, l0.CreateAccount(ctx, "a0"))

	require.NoError(t, l0.Deposit(ctx, "a0", ledger.Money(100)))
	checkBalance("a0", 100)

	require.NoError(t, l0.Withdraw(ctx, "a0", ledger.Money(50)))
	checkBalance("a0", 50)

	require.ErrorIs(t, l0.Withdraw(ctx, "a0", ledger.Money(100)), ledger.ErrNoMoney)

	require.NoError(t, l0.CreateAccount(ctx, "a1"))

	require.NoError(t, l0.Transfer(ctx, "a0", "a1", ledger.Money(40)))
	checkBalance("a0", 10)
	checkBalance("a1", 40)

	require.ErrorIs(t, l0.Transfer(ctx, "a0", "a1", ledger.Money(50)), ledger.ErrNoMoney)
}

func testErroneousCases(t *testing.T, ctx context.Context, l0 ledger.Ledger) {
	checkBalance := func(account ledger.ID, amount ledger.Money) {
		b, err := l0.GetBalance(ctx, account)
		require.NoError(t, err)
		require.Equal(t, amount, b)
	}

	require.NoError(t, l0.CreateAccount(ctx, "b0"))
	checkBalance("b0", 0)

	require.NoError(t, l0.Deposit(ctx, "b0", ledger.Money(100)))
	checkBalance("b0", 100)

	require.ErrorIs(t, l0.Deposit(ctx, "b0", ledger.Money(-100)), ledger.ErrNegativeAmount)
	checkBalance("b0", 100)

	require.ErrorIs(t, l0.Withdraw(ctx, "b0", ledger.Money(-50)), ledger.ErrNegativeAmount)
	checkBalance("b0", 100)

	require.Error(t, l0.Transfer(ctx, "b0", "b999", ledger.Money(50)))
	checkBalance("b0", 100)

	require.Error(t, l0.Transfer(ctx, "b999", "b0", ledger.Money(50)))
	checkBalance("b0", 100)

	require.NoError(t, l0.CreateAccount(ctx, "b999"))
	require.NoError(t, l0.Deposit(ctx, "b999", ledger.Money(200)))
	checkBalance("b999", 200)

	require.ErrorIs(t, l0.Transfer(ctx, "b0", "b999", ledger.Money(-50)), ledger.ErrNegativeAmount)
	checkBalance("b0", 100)
	checkBalance("b999", 200)

	require.NoError(t, l0.Transfer(ctx, "b0", "b999", ledger.Money(50)))
	checkBalance("b0", 50)
	checkBalance("b999", 250)

	require.Error(t, l0.Deposit(ctx, "c0", ledger.Money(100)))
	require.Error(t, l0.Withdraw(ctx, "c0", ledger.Money(100)))

	_, err := l0.GetBalance(ctx, "c0")
	require.Error(t, err)
}

func testTransactions(t *testing.T, ctx context.Context, l0 ledger.Ledger) {
	const nAccounts = 10
	const initialBalance = 5

	accounts := createTestAccounts(t, ctx, l0, nAccounts, initialBalance)

	var wg sync.WaitGroup
	done := make(chan struct{})

	spawnConcurrentOperations(t, ctx, l0, accounts, &wg, done)

	time.Sleep(time.Second * 10)
	close(done)
	wg.Wait()

	verifyTotalBalance(t, ctx, l0, accounts, nAccounts*initialBalance)
}

func createTestAccounts(t *testing.T, ctx context.Context, l0 ledger.Ledger, n int, balance ledger.Money) []ledger.ID {
	var accounts []ledger.ID
	for i := 0; i < n; i++ {
		id := ledger.ID(fmt.Sprint(i))
		accounts = append(accounts, id)

		require.NoError(t, l0.CreateAccount(ctx, id))
		require.NoError(t, l0.Deposit(ctx, id, balance))
	}
	return accounts
}

func spawnConcurrentOperations(t *testing.T, ctx context.Context, l0 ledger.Ledger, accounts []ledger.ID, wg *sync.WaitGroup, done chan struct{}) {
	for i := 0; i < len(accounts); i++ {
		account := accounts[i]
		next := accounts[(i+1)%len(accounts)]
		prev := accounts[(i+len(accounts)-1)%len(accounts)]

		spawnBalanceChecker(t, ctx, l0, account, wg, done)
		spawnTransfer(t, ctx, l0, account, next, wg, done)
		spawnTransfer(t, ctx, l0, account, prev, wg, done)
	}
}

func spawnBalanceChecker(t *testing.T, ctx context.Context, l0 ledger.Ledger, account ledger.ID, wg *sync.WaitGroup, done chan struct{}) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				balance, err := l0.GetBalance(ctx, account)
				if err != nil {
					if !errors.Is(err, ledger.ErrNoMoney) {
						t.Errorf("operation failed: %v", err)
						return
					}
					continue
				}
				if balance < 0 {
					t.Errorf("%q balance is negative", account)
					return
				}
			}
		}
	}()
}

func spawnTransfer(t *testing.T, ctx context.Context, l0 ledger.Ledger, from, to ledger.ID, wg *sync.WaitGroup, done chan struct{}) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				err := l0.Transfer(ctx, from, to, 1)
				if err != nil && !errors.Is(err, ledger.ErrNoMoney) {
					t.Errorf("operation failed: %v", err)
					return
				}
			}
		}
	}()
}

func verifyTotalBalance(t *testing.T, ctx context.Context, l0 ledger.Ledger, accounts []ledger.ID, expected int) {
	var total ledger.Money
	for _, account := range accounts {
		amount, err := l0.GetBalance(ctx, account)
		require.NoError(t, err)
		total += amount
	}
	require.Equal(t, ledger.Money(expected), total)
}
