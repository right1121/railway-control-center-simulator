package simulation

type Line struct {
	stations   []StationID
	blocks     []BlockID
	blockIndex map[string]int
}

func NewLine(stations []StationID, blocks []BlockID) (*Line, error) {
	if len(blocks) == 0 {
		return nil, ErrLineHasNoBlocks
	}
	if len(stations) != len(blocks)+1 {
		return nil, ErrLineStationsBlocksMismatch
	}

	stationSeen := make(map[string]struct{}, len(stations))
	for _, station := range stations {
		key := station.String()
		if _, exists := stationSeen[key]; exists {
			return nil, ErrLineDuplicateStationID
		}
		stationSeen[key] = struct{}{}
	}

	blockSeen := make(map[string]struct{}, len(blocks))
	blockIndex := make(map[string]int, len(blocks))
	for i, block := range blocks {
		key := block.String()
		if _, exists := blockSeen[key]; exists {
			return nil, ErrLineDuplicateBlockID
		}
		blockSeen[key] = struct{}{}
		blockIndex[key] = i
	}

	stationsCopy := make([]StationID, len(stations))
	copy(stationsCopy, stations)

	blocksCopy := make([]BlockID, len(blocks))
	copy(blocksCopy, blocks)

	return &Line{
		stations:   stationsCopy,
		blocks:     blocksCopy,
		blockIndex: blockIndex,
	}, nil
}

func (l *Line) Stations() []StationID {
	out := make([]StationID, len(l.stations))
	copy(out, l.stations)
	return out
}

func (l *Line) Blocks() []BlockID {
	out := make([]BlockID, len(l.blocks))
	copy(out, l.blocks)
	return out
}

func (l *Line) BlockAt(index int) (BlockID, bool) {
	if index < 0 || index >= len(l.blocks) {
		return BlockID{}, false
	}
	return l.blocks[index], true
}

func (l *Line) IndexOfBlock(id BlockID) (int, bool) {
	i, ok := l.blockIndex[id.String()]
	return i, ok
}

func (l *Line) NextBlock(id BlockID, forward bool) (BlockID, bool, error) {
	index, ok := l.IndexOfBlock(id)
	if !ok {
		return BlockID{}, false, ErrBlockNotFound
	}

	next := index + 1
	if !forward {
		next = index - 1
	}
	if next < 0 || next >= len(l.blocks) {
		return BlockID{}, false, nil
	}
	return l.blocks[next], true, nil
}
