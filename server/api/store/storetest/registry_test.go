package storetest

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type registration struct {
	group string
	test  string
}

func testName(test TestFunc) string {
	if test == nil {
		return "a nil entry"
	}

	qualified := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()

	return qualified[strings.LastIndex(qualified, ".")+1:]
}

func registrations(groups []Group) []registration {
	registered := make([]registration, 0)

	for _, group := range groups {
		for _, test := range group.Tests {
			registered = append(registered, registration{group: group.Name, test: testName(test)})
		}
	}

	return registered
}

func definedTestNames() []string {
	suite := reflect.TypeFor[*Suite]()

	defined := make([]string, 0, suite.NumMethod())

	for method := range suite.Methods() {
		if strings.HasPrefix(method.Name, "Test") {
			defined = append(defined, method.Name)
		}
	}

	return defined
}

func unregisteredTests(defined []string, registered []registration) []string {
	grouped := make(map[string]struct{}, len(registered))
	for _, entry := range registered {
		grouped[entry.test] = struct{}{}
	}

	complaints := make([]string, 0)

	for _, name := range defined {
		if _, ok := grouped[name]; !ok {
			complaints = append(complaints, name+" is defined on the suite and belongs to no group")
		}
	}

	return complaints
}

func staleRegistrations(defined []string, registered []registration) []string {
	onTheSuite := make(map[string]struct{}, len(defined))
	for _, name := range defined {
		onTheSuite[name] = struct{}{}
	}

	complaints := make([]string, 0)

	for _, entry := range registered {
		if _, ok := onTheSuite[entry.test]; !ok {
			complaints = append(complaints, entry.group+" registers "+entry.test+", which the suite does not define")
		}
	}

	return complaints
}

func duplicatedRegistrations(registered []registration) []string {
	order := make([]string, 0, len(registered))
	groupsByTest := make(map[string][]string, len(registered))

	for _, entry := range registered {
		if _, seen := groupsByTest[entry.test]; !seen {
			order = append(order, entry.test)
		}

		groupsByTest[entry.test] = append(groupsByTest[entry.test], entry.group)
	}

	complaints := make([]string, 0)

	for _, name := range order {
		if groups := groupsByTest[name]; len(groups) > 1 {
			complaints = append(complaints, name+" is registered more than once, in: "+strings.Join(groups, ", "))
		}
	}

	return complaints
}

// TestEveryStoreTestBelongsToExactlyOneGroup holds the registration table to the suite it claims to
// describe. A store test runs only because a group names it, and nothing else notices when one is
// written and never named, so this reads both sets of names and refuses any disagreement between
// them.
func TestEveryStoreTestBelongsToExactlyOneGroup(t *testing.T) {
	defined := definedTestNames()
	require.NotEmpty(t, defined, "the suite type exposes no test method, so every check below would pass vacuously")
	require.NotEmpty(t, Groups, "the registration table declares no group")

	registered := registrations(Groups)

	t.Run("unregisteredTests", func(t *testing.T) {
		assert.Empty(t, unregisteredTests(defined, registered))
	})

	t.Run("staleRegistrations", func(t *testing.T) {
		assert.Empty(t, staleRegistrations(defined, registered))
	})

	t.Run("duplicatedRegistrations", func(t *testing.T) {
		assert.Empty(t, duplicatedRegistrations(registered))
	})
}

// TestUnregisteredTestsCatchesATestInNoGroup drives the predicate with the failure it exists for: a
// test written on the suite that no group names, which today runs nowhere and reports nothing.
func TestUnregisteredTestsCatchesATestInNoGroup(t *testing.T) {
	complaints := unregisteredTests(
		[]string{"TestNamed", "TestForgotten"},
		[]registration{{group: "UserStore", test: "TestNamed"}},
	)

	assert.Equal(t, []string{"TestForgotten is defined on the suite and belongs to no group"}, complaints)
}

// TestStaleRegistrationsCatchesAGroupNamingATestTheSuiteDoesNotDefine drives the predicate with a
// registration a rename left behind. A method expression cannot go stale without the compiler
// saying so, so what this covers is a name the table acquired any other way.
func TestStaleRegistrationsCatchesAGroupNamingATestTheSuiteDoesNotDefine(t *testing.T) {
	complaints := staleRegistrations(
		[]string{"TestNamed"},
		[]registration{
			{group: "UserStore", test: "TestNamed"},
			{group: "DeviceStore", test: "TestRenamedAway"},
		},
	)

	assert.Equal(t, []string{"DeviceStore registers TestRenamedAway, which the suite does not define"}, complaints)
}

// TestDuplicatedRegistrationsCatchesATestInTwoGroups drives the predicate with the copy-paste that
// runs one test against two databases and hides which of them it needed.
func TestDuplicatedRegistrationsCatchesATestInTwoGroups(t *testing.T) {
	complaints := duplicatedRegistrations([]registration{
		{group: "UserStore", test: "TestNamed"},
		{group: "DeviceStore", test: "TestCopied"},
		{group: "SessionStore", test: "TestCopied"},
		{group: "TagStore", test: "TestTwiceOver"},
		{group: "TagStore", test: "TestTwiceOver"},
	})

	assert.Equal(t, []string{
		"TestCopied is registered more than once, in: DeviceStore, SessionStore",
		"TestTwiceOver is registered more than once, in: TagStore, TagStore",
	}, complaints)
}

// TestDefinedTestNamesRefusesEverythingButSuiteTests pins the collector, which is the one part that
// can fail by finding nothing and thereby make every predicate above pass.
func TestDefinedTestNamesRefusesEverythingButSuiteTests(t *testing.T) {
	defined := definedTestNames()

	assert.Contains(t, defined, "TestUserList")
	assert.NotContains(t, defined, "CreateUser")
}

// TestRegistrationsNamesEveryTestItWalks pins the name recovery the predicates report with: the
// table holds function values, and a complaint that cannot name its offender is unactionable.
func TestRegistrationsNamesEveryTestItWalks(t *testing.T) {
	registered := registrations([]Group{
		{Name: "UserStore", Tests: []TestFunc{(*Suite).TestUserList, (*Suite).TestUserResolve, nil}},
	})

	assert.Equal(t, []registration{
		{group: "UserStore", test: "TestUserList"},
		{group: "UserStore", test: "TestUserResolve"},
		{group: "UserStore", test: "a nil entry"},
	}, registered)
}
