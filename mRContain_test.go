package main

import (
	"testing"

	"github.com/miekg/dns"
)

type MRContainTest struct {
	msg  *dns.Msg
	conf *MRContain
	ok   bool
}

var testData = []MRContainTest{}

func TestProcess(t *testing.T) {
	/*for _, test := range testData {
		if output := process(test.arg1, test.arg2); output != test.expected {
			t.Errorf("Output %q not equal to expected %q", output, test.expected)
		}
	}*/
}
