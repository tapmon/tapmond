package mons

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"runtime"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type Mon struct {
	Id     []byte
	Scores []uint8
	Types  []string
	Level  int
	Nonce  int
}

// String returns a string representation of the monster.
func (m *Mon) String() string {
	return fmt.Sprintf("ID: %x, Scores: %v,  Level: %d, Nonce: %d",
		m.Id, m.Scores, m.Level, m.Nonce)
}

func GenerateMonster(blockHash, txHash *chainhash.Hash) (*Mon, error) {
	combinedHash, err := combineAndHash(blockHash, txHash)
	if err != nil {
		return nil, err
	}

	rarities := determineScores(combinedHash)

	monster := &Mon{
		Id:     combinedHash,
		Scores: rarities,
	}

	return monster, nil
}

func (m *Mon) GetLevelNonce(ctx context.Context, targetLevel,
	startNonce int) int {

	// Define difficulty as the target level (e.g., number of leading zeros)
	difficulty := targetLevel
	target := ""
	for i := 0; i < difficulty; i++ {
		target += "0"
	}

	numWorkers := runtime.NumCPU()
	var (
		wg           sync.WaitGroup
		found        bool
		mu           sync.Mutex
		levelUpNonce int
	)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID, startNonce int) {
			defer wg.Done()
			nonce := startNonce
			for {
				select {
				case <-ctx.Done():
					fmt.Printf("Context cancelled. Last nonce tried by worker %d: %d\n", workerID, nonce)
					return
				default:
					data := fmt.Sprintf("%s%d", m.Id, nonce)
					hash := sha256.Sum256([]byte(data))
					hashStr := hex.EncodeToString(hash[:])

					if hashStr[:difficulty] == target {
						mu.Lock()
						if !found {
							found = true
							m.Level = targetLevel
							levelUpNonce = nonce
							fmt.Printf("Monster leveled up! New level: %d, Nonce: %d, Hash: %s\n", m.Level, nonce, hashStr)
						}
						mu.Unlock()
						break
					}

					nonce += numWorkers
					mu.Lock()
					if found {
						mu.Unlock()
						break
					}
					mu.Unlock()
				}

			}
		}(i, i)
	}

	wg.Wait()
	return levelUpNonce
}

// VerifyLevelUp verifies that the level-up was performed correctly by checking
// if the hash of the combined hash and nonce produces a hash with the required
// number of leading zeros.
func (m *Mon) VerifyLevelUp(targetLevel int, nonce int) bool {
	difficulty := targetLevel
	target := ""
	for i := 0; i < difficulty; i++ {
		target += "0"
	}

	data := fmt.Sprintf("%s%d", m.Id, nonce)
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	return hashStr[:difficulty] == target
}

func combineAndHash(blockHash, txHash *chainhash.Hash) ([]byte, error) {
	combined := append(blockHash[:], txHash[:]...)
	hashedCombined := sha256.Sum256(combined)
	return hashedCombined[:], nil
}

// determineScores determines the rarity of each of the 32 attributes based on each byte of the combined hash.
func determineScores(combinedHash []byte) []uint8 {
	rarities := make([]uint8, len(combinedHash))
	for i := 0; i < len(combinedHash); i++ {
		rarities[i] = uint8(combinedHash[i])
	}
	return rarities
}

// CalculateRarityScore calculates the rarity score of the monster based on the first n attributes.
// The score is scaled to be between 0 and 1.
func (m *Mon) CalculateRarityScore(index int) float64 {
	if index > len(m.Scores) || index <= 0 {
		index = len(m.Scores)
	}
	score := 0.0
	for i := 0; i < index; i++ {
		score += float64(m.Scores[i])
	}
	return score / (255.0 * float64(index))
}

// func shouldHaveTwoTypes(rng *drng.DRNG, rarity Rarity) bool {
// 	chance := 0.0
// 	switch rarity {
// 	case Legendary:
// 		chance = 0.90
// 	case Epic:
// 		chance = 0.70
// 	case Rare:
// 		chance = 0.50
// 	case Normal:
// 		chance = 0.30
// 	}
// 	return rng.Float64() < chance
// }

// func generateTypes(rng *drng.DRNG, rarity Rarity) []string {
// 	firstTypeIndex := rng.Intn(len(types))
// 	firstType := types[firstTypeIndex]

// 	if shouldHaveTwoTypes(rng, rarity) {
// 		secondTypeIndex := rng.Intn(len(types))
// 		for secondTypeIndex == firstTypeIndex {
// 			secondTypeIndex = rng.Intn(len(types))
// 		}
// 		secondType := types[secondTypeIndex]
// 		return []string{firstType, secondType}
// 	}

// 	return []string{firstType}
// }

// func generateAttribute(rng *drng.DRNG, rarity Rarity) int {
// 	min, max := rollAttributeRange(rng, rarity)
// 	return rng.Intn(max-min+1) + min
// }

// func rollAttributeRange(rng *drng.DRNG, rarity Rarity) (int, int) {
// 	baseMin, baseMax := 50, 100
// 	switch rarity {
// 	case Rare:
// 		baseMin, baseMax = 60, 110
// 	case Epic:
// 		baseMin, baseMax = 70, 120
// 	case Legendary:
// 		baseMin, baseMax = 80, 130
// 	}
// 	min := rng.Intn(baseMax-baseMin+1) + baseMin
// 	max := rng.Intn(baseMax-min+1) + min
// 	return min, max
// }
