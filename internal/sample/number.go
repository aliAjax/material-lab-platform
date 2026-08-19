package sample

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var numberPattern = regexp.MustCompile(`^SMP-(\d{8})-(\d{5})$`)

func FormatNumber(day time.Time, sequence int64) (string, error) {
	return FormatNumberContext(context.Background(), day, sequence)
}

func FormatNumberContext(ctx context.Context, day time.Time, sequence int64) (string, error) {
	if sequence < 1 || sequence > 99999 {
		return "", fmt.Errorf("sequence must be between 1 and 99999")
	}
	return fmt.Sprintf("SMP-%s-%05d", day.UTC().Format("20060102"), sequence), nil
}

func ParseNumber(number string) (time.Time, int64, error) {
	return ParseNumberContext(context.Background(), number)
}

func ParseNumberContext(ctx context.Context, number string) (time.Time, int64, error) {
	parts := numberPattern.FindStringSubmatch(number)
	if parts == nil {
		return time.Time{}, 0, fmt.Errorf("invalid sample number")
	}
	day, err := time.Parse("20060102", parts[1])
	if err != nil {
		return time.Time{}, 0, err
	}
	sequence, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return time.Time{}, 0, err
	}
	return day, sequence, nil
}
