// EventRule.g4
grammar EventRule;

// Tokens
AND: 'and';
OR: 'or';
NOT: 'not';
EQU: '=';
NEQ: '!=';
GT: '>';
LT: '<';
GTE: '>=';
LTE: '<=';
CONTAINS: 'contains';
IN: 'in';
COMMA: ',';
NUMBER: [-]?[0-9]+('.'[0-9]+)?;
VAR: [a-zA-Z0-9_.-]+;
STRING: '"' (ESC|.)*? '"';
WHITESPACE: [ \t\r\n] ->skip;

fragment
ESC: '\\"' | '\\\\';

// Rules
start
   : expression EOF
   ;

expression
   : expression op=(AND|OR) expression      # AndOr
   | NOT expression                         # Not
   | '(' expression ')'                     # Parenthesis
   | VAR op=(EQU|CONTAINS) STRING           # StringEqualContains
   | VAR IN '(' STRING (COMMA STRING)* ')'  # StringIn
   | VAR op=(EQU|NEQ|GT|LT|GTE|LTE) NUMBER  # CompareNumber
   | VAR                                    # Variable
   ;
