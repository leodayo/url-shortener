package randstr

import "crypto/rand"

const validCharacters = "abcdefghjkmnpqrstuvwxyz123456789"

func RandString(n int) (string, error) {
	output := make([]byte, n)
	randIndices := make([]byte, n)
	_, err := rand.Read(randIndices)

	if err != nil {
		return "", err
	}

	for i := range output {
		output[i] = validCharacters[int(randIndices[i])%len(validCharacters)]
	}

	return string(output), nil
}
