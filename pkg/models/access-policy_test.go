package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecisionMessage(t *testing.T) {
	cases := []struct {
		description     string
		decision        Decision
		expectedMessage string
	}{
		{
			description:     "an allowed decision has no denial message",
			decision:        Decision{Allowed: true},
			expectedMessage: "",
		},
		{
			description:     "not a member",
			decision:        Decision{Reason: ReasonNotAMember},
			expectedMessage: "user is not a member of the namespace",
		},
		{
			description:     "denied by a policy names the policy",
			decision:        Decision{Reason: ReasonDeniedByPolicy, PolicyName: "block contractors"},
			expectedMessage: `denied by policy "block contractors"`,
		},
		{
			description:     "an unevaluable policy names the policy",
			decision:        Decision{Reason: ReasonPolicyUnevaluable, PolicyName: "block contractors"},
			expectedMessage: `denied: policy "block contractors" could not be evaluated`,
		},
		{
			description:     "no grant names the requested login",
			decision:        Decision{Reason: ReasonNoGrant, Login: "root"},
			expectedMessage: `no policy grants "root" on this device`,
		},
		{
			description:     "a denial with no reason falls back to a generic sentence",
			decision:        Decision{},
			expectedMessage: "denied by the access policies",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			require.Equal(t, tc.expectedMessage, tc.decision.Message())
		})
	}
}
