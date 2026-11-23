package calculator

type Problem struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (p *Problem) Error() string {
	return p.Message
}

func SyntaxProblem(message string) *Problem {
	return &Problem{400, "invalid_expression", message}
}

var (
	RangeProblem = &Problem{422, "out_of_range", "Number exceeds the supported range."}
	ZeroDivisor  = &Problem{422, "division_by_zero", "Cannot divide by zero."}
)
