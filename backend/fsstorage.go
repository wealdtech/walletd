package backend

import "fmt"

type FSStorage struct {
	// Currently in-memory.
	accountStates     map[string]*State
	attestationStates map[string]*State
	proposalStates    map[string]*State
}

// NewFSStorage creates a new filesystem storage.
func NewFSStorage(base string) (*FSStorage, error) {
	return &FSStorage{
		accountStates:     make(map[string]*State),
		attestationStates: make(map[string]*State),
		proposalStates:    make(map[string]*State),
	}, nil
}

// FetchBeaconProposalState fetches beacon proposal state for a given key.
func (s *FSStorage) FetchBeaconProposalState(pubKey []byte) (*State, error) {
	key := fmt.Sprintf("beaconproposal-%0x", pubKey)
	var state *State
	var exists bool
	if state, exists = s.proposalStates[key]; !exists {
		state = NewState()
		s.proposalStates[key] = state
	}
	return state, nil
}

// StoreBeaconProposalState stores beacon proposal state for a given key.
func (s *FSStorage) StoreBeaconProposalState(pubKey []byte, state *State) error {
	key := fmt.Sprintf("beaconproposal-%0x", pubKey)
	s.proposalStates[key] = state
	return nil
}

// FetchBeaconAttestationState fetches beacon attestation state for a given key.
func (s *FSStorage) FetchBeaconAttestationState(pubKey []byte) (*State, error) {
	key := fmt.Sprintf("beaconattestation-%0x", pubKey)
	var state *State
	var exists bool
	if state, exists = s.attestationStates[key]; !exists {
		state = NewState()
		s.attestationStates[key] = state
	}
	return state, nil
}

// StoreBeaconAttestationState stores beacon attestation state for a given key.
func (s *FSStorage) StoreBeaconAttestationState(pubKey []byte, state *State) error {
	key := fmt.Sprintf("beaconattestation-%0x", pubKey)
	s.attestationStates[key] = state
	return nil
}

// FetchListAccountsState fetches list accounts state for a given key.
func (s *FSStorage) FetchListAccountsState(pubKey []byte) (*State, error) {
	key := fmt.Sprintf("listaccouts-%0x", pubKey)
	var state *State
	var exists bool
	if state, exists = s.accountStates[key]; !exists {
		state = NewState()
		s.accountStates[key] = state
	}
	return state, nil
}

// StoreListAccountsState stores list accounts state for a given key.
func (s *FSStorage) StoreListAccountsState(pubKey []byte, state *State) error {
	key := fmt.Sprintf("listaccouts-%0x", pubKey)
	s.accountStates[key] = state
	return nil
}
