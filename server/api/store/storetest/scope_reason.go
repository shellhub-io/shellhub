package storetest

// reasonTestQueryMechanics is the reason carried by store tests that exercise a query's mechanics
// rather than its namespace bound. It is one constant so that grepping NewUnbounded for the
// production audit surface is not drowned out by the suites repeating the same literal.
const reasonTestQueryMechanics = "test: exercising the query itself, not its namespace bound"
