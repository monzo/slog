package slog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountFmtOperands(t *testing.T) {
	cases := map[string]int{
		`%%`:    0,
		`%%s`:   0,
		`%v`:    1,
		`%#v`:   1,
		`%T`:    1,
		`%t`:    1,
		`%c`:    1,
		`%d`:    1,
		`%o`:    1,
		`%O`:    1,
		`%U`:    1,
		`%b`:    1,
		`%e`:    1,
		`%E`:    1,
		`%f`:    1,
		`%F`:    1,
		`%g`:    1,
		`%G`:    1,
		`%s`:    1,
		`%q`:    1,
		`%x`:    1,
		`%X`:    1,
		`%p`:    1,
		`%9f`:   1,
		`%.2f`:  1,
		`%9.2f`: 1,
		`%9.f`:  1,
		`% d`:   1,
		`%09d`:  1,

		`%%s %s %s`:                        2,
		`%6.2f`:                            1,
		`%d %d %#[1]x %#x`:                 2,
		`%d %d %d %[1]d %d %[3]x %d %d %x`: 6,
		`%s %% %%s %s`:                     2,
		`%s %s`:                            2,
		`%s %s %d`:                         3,
		`%[2]d %[1]d`:                      2,
		`%[3]*.[2]*[1]f`:                   3,
		`%[3]*.[2]*[1]f %[3]*.[2]*[1]f %s`: 3}

	for input, count := range cases {
		assert.Equal(t, count, countFmtOperands(input), input)
	}
}
