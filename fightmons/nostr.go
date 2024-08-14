package fightmons

import (
	"encoding/json"

	"github.com/nbd-wtf/go-nostr"
)

const (
	// FightMonKindStartRange is the start of the range of kinds for FightMon
	// events.
	FightMonKindStartRange = 86061

	// FightMonLookingForMatch is the kind for FightMon events where a player
	// is looking for a match.
	FightMonLookingForMatch = FightMonKindStartRange + iota

	// FightMonRequestMatch is the kind for FightMon events where a player
	// requests a match from a player looking for a match.
	FightMonRequestMatch

	// FightMonAcceptMatch is the kind for FightMon events where a player
	// accepts a match request.
	FightMonAcceptMatch

	// FightMonStartMatch is the kind for FightMon events where a match is
	// started.
	FightMonStartMatch

	// FightMonRound is the kind for FightMon events where a player submits
	// their move for a round.
	FightMonRound

	// FightMonRoundCommit is the kind for FightMon events where a player
	// commits their move for a round by publishing their rng seed.
	FightMonRoundCommit
)

func GetFightMonLookingForMatchEvent(matchId string) *nostr.Event {
	return &nostr.Event{
		Kind: FightMonLookingForMatch,
		Tags: nostr.Tags{
			nostr.Tag{"match_id", matchId},
		},
	}
}

func GetFightMonRequestMatchEvent(matchId string, fightmon FightMon,
) (*nostr.Event, error) {

	data, err := json.Marshal(fightmon)
	if err != nil {
		return nil, err
	}

	return &nostr.Event{
		Kind: FightMonRequestMatch,
		Tags: nostr.Tags{
			nostr.Tag{"match_id", matchId},
		},
		Content: string(data),
	}, nil
}

func GetFightMonAcceptMatchEvent(matchId string, fightMons FightMon,
) (*nostr.Event, error) {

	data, err := json.Marshal(fightMons)
	if err != nil {
		return nil, err
	}

	return &nostr.Event{
		Kind: FightMonAcceptMatch,
		Tags: nostr.Tags{
			nostr.Tag{"match_id", matchId},
		},
		Content: string(data),
	}, nil
}

func GetFightMonStartMatchEvent(matchId string) *nostr.Event {
	return &nostr.Event{
		Kind: FightMonStartMatch,
		Tags: nostr.Tags{
			nostr.Tag{"match_id", matchId},
		},
	}
}

type FightMonRoundEvent struct {
	RoundId int    `json:"round_id"`
	Action  int    `json:"action"`
	RngHash string `json:"rng_hash"`
}

func GetFightMonRoundEvent(matchId string, round FightMonRoundEvent,
) (*nostr.Event, error) {

	data, err := json.Marshal(round)
	if err != nil {
		return nil, err
	}

	return &nostr.Event{
		Kind: FightMonRound,
		Tags: nostr.Tags{
			nostr.Tag{"match_id", matchId},
		},
		Content: string(data),
	}, nil
}

type FightMonRoundCommitEvent struct {
	RoundID int    `json:"round_id"`
	Action  int    `json:"action"`
	RngSeed string `json:"rng_seed"`
}

func GetFightMonRoundCommitEvent(matchId string, commit FightMonRoundCommitEvent,
) (*nostr.Event, error) {

	data, err := json.Marshal(commit)
	if err != nil {
		return nil, err
	}

	return &nostr.Event{
		Kind: FightMonRoundCommit,
		Tags: nostr.Tags{
			nostr.Tag{"match_id", matchId},
		},
		Content: string(data),
	}, nil
}
