export type Calculation = { expression: string; value: string };
export type CalculationBody = { expression: string };
export type Submission = { body: CalculationBody; repeat: string[] | null };

export type State = {
  tokens: string[];
  value: string;
  finished: boolean;
  repeat: string[] | null;
  request: Submission | null;
  error: string;
  failedExpression: string;
  history: Calculation[];
};

export type Reply =
  | { type: 'success'; request: Submission; result: number }
  | { type: 'failure'; request: Submission; message: string };
