package main

import "testing"

func TestPrepAnswer(t *testing.T) {
	cases := []struct {
		want   string
		answer string
	}{
		{
			"nigel thornberry",
			"Who is Nigel Thornberry?",
		},
		{
			"Pisces",
			"What is Pisces?",
		},
		{
			"a catsuit",
			"What is a catsuit?",
		},
	}

	for _, c := range cases {
		got := prepAnswer(c.answer)
		if got != c.want {
			t.Fatalf("prepAnswer(%s) == %s, want %s", c.answer, got, c.want)
		}
	}
}
