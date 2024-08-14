package mons

import (
	"context"
	"crypto/rand"
	"runtime"
	"sync"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/require"
)

var (
	testBlockHash = chainhash.Hash{
		0x03,
	}
	testTxHash = chainhash.Hash{
		0x10,
	}
)

func TestGenerateMon(t *testing.T) {
	monster, err := GenerateMonster(&testBlockHash, &testTxHash)
	require.NoError(t, err)
	t.Logf("Generated monster: %v", monster)

	monster, err = GenerateMonster(&testTxHash, &testBlockHash)
	require.NoError(t, err)
	t.Logf("Generated monster: %v", monster)

	score := monster.CalculateRarityScore(0)
	t.Logf("Rarity score: %v", score)
}

func TestFindHighRarityMon(t *testing.T) {
	for i := 0; i < 100000; i++ {
		// randomize test tx hash
		testTxHash = getRandomHash()

		monster, err := GenerateMonster(&testBlockHash, &testTxHash)
		require.NoError(t, err)

		rarityScore := monster.CalculateRarityScore(0)
		//t.Logf("Rarity score: %v", rarityScore)
		if rarityScore > 75.0 {
			t.Logf("Found high rarity %v monster: %v at %x", rarityScore, monster, testTxHash)
		}
	}
}

func TestFindHighRarityMon2(t *testing.T) {
	numWorkers := runtime.NumCPU() // Use the number of CPUs available
	const numTasks = 100000000
	tasksPerWorker := numTasks / numWorkers

	var wg sync.WaitGroup

	// Worker function
	worker := func(start, count int) {
		defer wg.Done() // Mark this worker as done when it returns
		for i := start; i < start+count; i++ {
			// Randomize test tx hash
			testTxHash := getRandomHash()

			monster, err := GenerateMonster(&testBlockHash, &testTxHash)
			require.NoError(t, err)

			rarityScore := monster.CalculateRarityScore(0)
			if rarityScore > 0.8 {
				t.Logf("Found high rarity %v monster: %v at %x", rarityScore, monster, testTxHash)
			}
		}
	}

	// Launch workers
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		start := i * tasksPerWorker
		count := tasksPerWorker
		if i == numWorkers-1 {
			// Last worker takes any remaining tasks
			count = numTasks - start
		}
		go worker(start, count)
	}

	// Wait for all workers to finish
	wg.Wait()
}

// getRandomHash returns a random hash for testing purposes.
func getRandomHash() chainhash.Hash {
	var hash chainhash.Hash
	_, err := rand.Read(hash[:])
	if err != nil {
		panic(err)
	}
	return hash
}

func TestLevel(t *testing.T) {
	ctxt := context.Background()
	monster, err := GenerateMonster(&testBlockHash, &testTxHash)
	require.NoError(t, err)
	t.Logf("Generated monster: %v", monster)
	monster.GetLevelNonce(ctxt, 1, 0)
	require.Equal(t, 1, monster.Level)

	lvelUpNonce := monster.GetLevelNonce(ctxt, 7, 0)
	require.Equal(t, 7, monster.Level)
	require.True(t, monster.VerifyLevelUp(5, lvelUpNonce))
}
