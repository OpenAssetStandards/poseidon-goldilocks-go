package pgoldilocks

import (
	"math/big"

	ffg "github.com/OpenAssetStandards/poseidon-goldilocks-go/ffg"
)

func zero() *ffg.Element {
	return ffg.NewElement()
}

// exp7 performs x^7 mod p
func exp7(a *ffg.Element) {
	a.Exp(*a, big.NewInt(7)) //nolint:gomnd
}

// exp7state perform exp7 for whole state
func exp7state(state []*ffg.Element) {
	for i := 0; i < len(state); i++ {
		exp7(state[i])
	}
}

// ark computes Add-Round Key, from the paper
func ark(state []*ffg.Element, it int) {
	for i := 0; i < len(state); i++ {
		state[i].Add(state[i], C[it+i])
	}
}

// mix returns [[matrix]] * [vector]
func mix(state []*ffg.Element, opt bool) []*ffg.Element {
	mul := zero()
	newState := make([]*ffg.Element, mLen)
	for i := 0; i < mLen; i++ {
		newState[i] = zero()
	}
	for i := 0; i < mLen; i++ {
		newState[i].SetUint64(0)
		for j := 0; j < mLen; j++ {
			if opt {
				mul.Mul(P[j][i], state[j])
			} else {
				mul.Mul(M[j][i], state[j])
			}
			newState[i].Add(newState[i], mul)
		}
	}
	return newState
}
func HashState(state []*ffg.Element) ([]*ffg.Element, error) {

	for i := 0; i < mLen; i++ {
		state[i].Add(state[i], C[i])
	}

	for r := 0; r < NROUNDSF/2; r++ {
		exp7state(state)
		ark(state, (r+1)*mLen)
		state = mix(state, r == NROUNDSF/2-1)
	}

	for r := 0; r < NROUNDSP; r++ {
		exp7(state[0])
		state[0].Add(state[0], C[(NROUNDSF/2+1)*mLen+r])

		s0 := zero()
		mul := zero()
		mul.Mul(S[(mLen*2-1)*r], state[0])
		s0.Add(s0, mul)
		for i := 1; i < mLen; i++ {
			mul.Mul(S[(mLen*2-1)*r+i], state[i])
			s0.Add(s0, mul)
			mul.Mul(S[(mLen*2-1)*r+mLen+i-1], state[0])
			state[i].Add(state[i], mul)
		}
		state[0] = s0
	}

	for r := 0; r < NROUNDSF/2; r++ {
		exp7state(state)
		if r < NROUNDSF/2-1 {
			ark(state, (NROUNDSF/2+1+r)*mLen+NROUNDSP)
		}

		state = mix(state, false)
	}
	return state, nil
}
func HashNToMNoPad(inputs []*ffg.Element, numOutputs int) ([]*ffg.Element, error) {

	state := make([]*ffg.Element, mLen)
	for i := 0; i < mLen; i++ {
		state[i] = ffg.NewElement().SetZero()
	}
	nChunks := len(inputs) / SPONGE_RATE
	for i := 0; i < nChunks; i++ {

		state[0] = inputs[i*8+0]
		state[1] = inputs[i*8+1]
		state[2] = inputs[i*8+2]
		state[3] = inputs[i*8+3]
		state[4] = inputs[i*8+4]
		state[5] = inputs[i*8+5]
		state[6] = inputs[i*8+6]
		state[7] = inputs[i*8+7]
		ns, err := HashState(state)
		if err != nil {
			return nil, err
		}
		state[0] = ns[0]
		state[1] = ns[1]
		state[2] = ns[2]
		state[3] = ns[3]
		state[4] = ns[4]
		state[5] = ns[5]
		state[6] = ns[6]
		state[7] = ns[7]
		state[8] = ns[8]
		state[9] = ns[9]
		state[10] = ns[10]
		state[11] = ns[11]
	}
	start := nChunks * SPONGE_RATE
	remaining := len(inputs) - start
	if remaining > 0 && remaining < mLen {
		for i := 0; i < remaining; i++ {
			state[i] = inputs[start+i]
		}
		ns, err := HashState(state)
		if err != nil {
			return nil, err
		}
		state[0] = ns[0]
		state[1] = ns[1]
		state[2] = ns[2]
		state[3] = ns[3]
		state[4] = ns[4]
		state[5] = ns[5]
		state[6] = ns[6]
		state[7] = ns[7]
		state[8] = ns[8]
		state[9] = ns[9]
		state[10] = ns[10]
		state[11] = ns[11]
	}
	mOut := make([]*ffg.Element, numOutputs)
	if numOutputs <= SPONGE_RATE {
		for i := 0; i < numOutputs; i++ {
			mOut[i] = state[i]
		}
		return mOut, nil
	} else {

		nOutputRounds := numOutputs / SPONGE_RATE
		if nOutputRounds*SPONGE_RATE < numOutputs {
			nOutputRounds = nOutputRounds + 1
		}
		outputsPushed := 0
		for i := 0; i < nOutputRounds; i++ {
			for x := 0; x < SPONGE_RATE; x++ {
				mOut[outputsPushed] = state[i]
				outputsPushed = outputsPushed + 1
				if outputsPushed == numOutputs {
					return mOut, nil
				}
			}
			ns, err := HashState(state)
			if err != nil {
				return nil, err
			}
			state = ns
		}
		return mOut, nil

	}

}
func HashNoPad(inputs []*ffg.Element) (*HashOut256, error) {
	result, err := HashNToMNoPad(inputs, 4)
	if err != nil {
		return nil, err
	}

	return &HashOut256{result[0].ToUint64Regular(), result[1].ToUint64Regular(),
		result[2].ToUint64Regular(), result[3].ToUint64Regular()}, nil
}
func HashPad(inputs []*ffg.Element) (*HashOut256, error) {
	lenInputs := len(inputs)
	padAmount := ((2 + lenInputs) % STATE_SIZE)
	totalSize := lenInputs + 2 + padAmount
	fullInputs := make([]*ffg.Element, totalSize)
	fullInputs[0] = ffg.NewElement().SetOne()
	for i := 0; i < lenInputs; i++ {
		fullInputs[i+1] = inputs[i]
	}
	for i := lenInputs + 1; i < (totalSize - 1); i++ {
		fullInputs[i] = ffg.NewElement().SetZero()
	}
	fullInputs[totalSize-1] = ffg.NewElement().SetOne()

	result, err := HashNToMNoPad(fullInputs, 4)
	if err != nil {
		return nil, err
	}

	return &HashOut256{result[0].ToUint64Regular(), result[1].ToUint64Regular(),
		result[2].ToUint64Regular(), result[3].ToUint64Regular()}, nil
}
func HashTwoToState(a *HashOut256, b *HashOut256,
	capacity [CAPLEN]uint64) ([]*ffg.Element, error) {
	state := make([]*ffg.Element, mLen)
	state[0] = ffg.NewElement().SetUint64(a[0])
	state[1] = ffg.NewElement().SetUint64(a[1])
	state[2] = ffg.NewElement().SetUint64(a[2])
	state[3] = ffg.NewElement().SetUint64(a[3])
	state[4] = ffg.NewElement().SetUint64(b[0])
	state[5] = ffg.NewElement().SetUint64(b[1])
	state[6] = ffg.NewElement().SetUint64(b[2])
	state[7] = ffg.NewElement().SetUint64(b[3])
	state[8] = ffg.NewElement().SetUint64(capacity[0])
	state[9] = ffg.NewElement().SetUint64(capacity[1])
	state[10] = ffg.NewElement().SetUint64(capacity[2])
	state[11] = ffg.NewElement().SetUint64(capacity[3])
	return HashState(state)
}

func HashTwoToOne(a *HashOut256, b *HashOut256) (*HashOut256, error) {
	//fmt.Printf("a: %s\n", a.ElementString())
	//fmt.Printf("b: %s\n", b.ElementString())
	result, err := HashTwoToState(a, b, [CAPLEN]uint64{0, 0, 0, 0})
	if err != nil {
		return nil, err
	}
	/*
		fmt.Print("result = [")
		for _, r := range result {
			fmt.Printf("%d, ", r.ToUint64Regular())
		}
		fmt.Print("]\n")
		fmt.Print("result2 = [")
		for _, r := range result {
			fmt.Printf("%s, ", r.String())
		}
		fmt.Print("]\n")
	*/
	return &HashOut256{result[0].ToUint64Regular(), result[1].ToUint64Regular(),
		result[2].ToUint64Regular(), result[3].ToUint64Regular()}, nil
}

func HashPadU64Array(input []uint64) (*HashOut256, error) {
	return HashPad(ffg.FromU64Array(input))
}
func HashNoPadU64Array(input []uint64) (*HashOut256, error) {
	return HashNoPad(ffg.FromU64Array(input))
}

func HashTwoToOneLeaf(a *HashOut256, b *HashOut256) (*HashOut256, error) {
	result, err := HashTwoToState(a, b, [CAPLEN]uint64{0, 0, 0, 0})
	if err != nil {
		return nil, err
	}

	result[0] = ffg.NewElement().SetUint64(1)
	result[1] = ffg.NewElement().SetUint64(1)
	result[2] = ffg.NewElement().SetUint64(0)
	result[3] = ffg.NewElement().SetUint64(1)

	result, err = HashState(result)
	if err != nil {
		return nil, err
	}

	return &HashOut256{result[0].ToUint64Regular(), result[1].ToUint64Regular(),
		result[2].ToUint64Regular(), result[3].ToUint64Regular()}, nil
}

type PoseidonGoldilocksHasher struct{}

func (PoseidonGoldilocksHasher) HashTwoToOne(a *HashOut256, b *HashOut256) (*HashOut256, error) {
	r, err := HashTwoToOne(a, b)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("HashTwoToOne(%s, %s) = %s\n", a.Hex(), b.Hex(), r.Hex())
	return r, err
}
func (PoseidonGoldilocksHasher) HashTwoToOneLeaf(a *HashOut256, b *HashOut256) (*HashOut256, error) {
	/*r, err := HashTwoToOneLeaf(a, b)
	if err != nil {
		return nil, err
	}
	fmt.Printf("HashTwoToOneLeaf(%s, %s) = %s\n", a.Hex(), b.Hex(), r.Hex())
	return r, err*/
	output := NewHashOut256PtrFromUint64(0)
	output.Copy(b)
	return output, nil
}
func (h PoseidonGoldilocksHasher) ComputeZeroHashes(levels uint64) ([]*HashOut256, error) {
	zeroHashes := make([]*HashOut256, levels)
	zeroHashes[levels-1] = NewHashOut256PtrFromUint64(0)
	for i := int(levels - 2); i >= 0; i-- {
		h, err := h.HashTwoToOne(zeroHashes[i+1], zeroHashes[i+1])
		if err != nil {
			return nil, err
		}
		zeroHashes[i] = h
	}
	return zeroHashes, nil
}

/*
func dbgPrintState(prefix string, inpBI [NROUNDSF]uint64, capBI
[CAPLEN]uint64) {
	fmt.Printf("[state-%s]\ninpBI=[", prefix)
	for i := 0; i < NROUNDSF; i++ {
		fmt.Printf("%d,", inpBI[i])
	}
	fmt.Print("]; capBI = [")
	for i := 0; i < CAPLEN; i++ {
		fmt.Printf("%d,", inpBI[i])
	}
	fmt.Print("]\n")
}

// Hash computes the hash for the given inputs
func HashToState(inpBI [NROUNDSF]uint64, capBI [CAPLEN]uint64)
([]*ffg.Element, error) {
	dbgPrintState("HashToState", inpBI, capBI)
	state := make([]*ffg.Element, mLen)
	for i := 0; i < NROUNDSF; i++ {
		state[i] = ffg.NewElement().SetUint64(inpBI[i])
	}
	for i := 0; i < CAPLEN; i++ {
		state[i+NROUNDSF] = ffg.NewElement().SetUint64(capBI[i])
	}
	newState, err := HashState(state)
	if err != nil {
		return state, err
	}
	return newState, err
}

// Hash computes the hash for the given inputs
func Hash(inpBI [NROUNDSF]uint64, capBI [CAPLEN]uint64) ([CAPLEN]uint64,
error) {
	newState, err := HashToState(inpBI, capBI)
	if err != nil {
		return [CAPLEN]uint64{0, 0, 0, 0}, err
	}

	return [CAPLEN]uint64{
		newState[0].ToUint64Regular(),
		newState[1].ToUint64Regular(),
		newState[2].ToUint64Regular(),
		newState[3].ToUint64Regular(),
	}, nil
}
func HashTwoToOne(a *big.Int, b *big.Int) ([CAPLEN]uint64, error) {
	abBytes := make([]byte, 64)
	a.FillBytes(abBytes[0:32])
	b.FillBytes(abBytes[32:64])
	fmt.Printf("hashtwotoone input %s\n", hex.EncodeToString(abBytes))
	return Hash([8]uint64{
		binary.BigEndian.Uint64(abBytes[56:64]),
		binary.BigEndian.Uint64(abBytes[48:56]),
		binary.BigEndian.Uint64(abBytes[40:48]),
		binary.BigEndian.Uint64(abBytes[32:40]),
		binary.BigEndian.Uint64(abBytes[24:32]),
		binary.BigEndian.Uint64(abBytes[16:24]),
		binary.BigEndian.Uint64(abBytes[8:16]),

		binary.BigEndian.Uint64(abBytes[0:8]),
	}, [4]uint64{
		0, 0, 0, 0,
	})

}

func HashTwoToOneLeaf(a *big.Int, b *big.Int) ([CAPLEN]uint64, error) {
	abBytes := make([]byte, 64)
	a.FillBytes(abBytes[0:32])
	b.FillBytes(abBytes[32:64])

	r1, err := HashToState([8]uint64{
		binary.BigEndian.Uint64(abBytes[56:64]),
		binary.BigEndian.Uint64(abBytes[48:56]),
		binary.BigEndian.Uint64(abBytes[40:48]),
		binary.BigEndian.Uint64(abBytes[32:40]),
		binary.BigEndian.Uint64(abBytes[24:32]),
		binary.BigEndian.Uint64(abBytes[16:24]),
		binary.BigEndian.Uint64(abBytes[8:16]),

		binary.BigEndian.Uint64(abBytes[0:8]),
	}, [4]uint64{
		0, 0, 0, 0,
	})
	if err != nil {
		return [CAPLEN]uint64{0, 0, 0, 0}, err
	}

	r1[0] = ffg.NewElement().SetUint64(1)
	r1[1] = ffg.NewElement().SetUint64(1)
	r1[2] = ffg.NewElement().SetUint64(0)
	r1[3] = ffg.NewElement().SetUint64(1)
	finalState, err := HashState(r1)
	if err != nil {
		return [CAPLEN]uint64{0, 0, 0, 0}, err
	}

	return [CAPLEN]uint64{
		finalState[0].ToUint64Regular(),
		finalState[1].ToUint64Regular(),
		finalState[2].ToUint64Regular(),
		finalState[3].ToUint64Regular(),
	}, nil
}
*/
