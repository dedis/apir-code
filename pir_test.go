package main

/*
func TestMultiBitVectorOneKbBytes(t *testing.T) {
	dbLen := oneKB
	blockLen := constants.BlockLength
	elemSize := field.Bits
	nRows := 1
	nCols := dbLen / (elemSize * blockLen * nRows)

	// functions defined in vpir_test.go
	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitBytes(xofDB, dbLen, nRows, blockLen)

	retrieveBlocksBytes(t, xof, db, nRows*nCols, "MultiBitVectorOneKb")
}

func TestMultiBitMatrixOneKbBytes(t *testing.T) {
	dbLen := oneKB
	blockLen := constants.BlockLength
	elemSize := field.Bits
	numBlocks := dbLen / (elemSize * blockLen)
	nCols := int(math.Sqrt(float64(numBlocks)))
	nRows := nCols

	// functions defined in vpir_test.go
	xofDB := getXof(t, "db key")
	xof := getXof(t, "client key")

	db := database.CreateRandomMultiBitBytes(xofDB, dbLen, nRows, blockLen)

	retrieveBlocksBytes(t, xof, db, numBlocks, "MultiBitMatrixOneKb")
}

func retrieveBlocksBytes(t *testing.T, rnd io.Reader, db *database.Bytes, numBlocks int, testName string) {
	c := client.NewPIR(rnd, &db.Info)
	s0 := server.NewPIR(db)
	s1 := server.NewPIR(db)

	totalTimer := monitor.NewMonitor()
	for i := 0; i < numBlocks; i++ {
		queries := c.Query(i, 2)

		a0 := s0.Answer(queries[0])
		a1 := s1.Answer(queries[1])

		answers := [][]byte{a0, a1}

		res, err := c.Reconstruct(answers)
		require.NoError(t, err)
		if db.BlockSize == constants.SingleBitBlockLength {
			require.Equal(t, db.Entries[i/db.NumColumns][i%db.NumColumns:i%db.NumColumns+1], res)
		} else {
			require.ElementsMatch(t, db.Entries[i/db.NumColumns][(i%db.NumColumns)*db.BlockSize:(i%db.NumColumns+1)*db.BlockSize], res)
		}

	}
	fmt.Printf("Total time %s: %.2fms\n", testName, totalTimer.Record())
}
*/
