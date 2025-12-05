package handler

import (
	"errors"
	"testing"
)

func TestExtractUserId(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected int64
		err      error
	}{
		{"Zero user id", "User ID: 0", 0, nil},
		{"Correct user id", "User ID: 1", 1, nil},
		{"Correct user id", "User ID: 1234", 1234, nil},
		{"Case insensetive user id", "user id: 1234567890", 1234567890, nil},
		{"Negative user id", "User Id: -1234567890", -1234567890, nil},
		{"No space user id", "UserId: 123", 123, nil},
		{"Dirty user id", "User Id: 123asd", 123, nil},
		{"Empty user id", "User ID:", 0, errors.New("regexp failed - userId not found")},
		{"Dirty user id", "User Id: as12", 0, errors.New("regexp failed - userId not found")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := extractUserId(tc.input)

			if actual != tc.expected {
				t.Errorf("extractUserId(%s) = %d; ожидалось %d", tc.input, actual, tc.expected)
			}
			if (tc.err == nil && err != tc.err) || (tc.err != nil && errors.Is(err, tc.err)) {
				t.Errorf("extractUserId(%s) вернул ошибку %v; ожидалось %v", tc.input, err, tc.err)
			}
		})
	}

}
