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
NOTCONTAINS: 'not contains';
IN: 'in';
NOTIN: 'not in';
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
   : expression op=(AND|OR) expression                                  # AndOr
   | NOT expression                                                     # Not
   | '(' expression ')'                                                 # Parenthesis
   | VAR op=(EQU|NEQ|GT|LT|GTE|LTE) (STRING|NUMBER)                     # Compare
   | VAR op=(CONTAINS|NOTCONTAINS) (STRING|NUMBER)                      # ContainsOrNot
   | VAR op=(IN|NOTIN) '(' (NUMBER|STRING) (COMMA (NUMBER|STRING))* ')' # InOrNot
   | VAR                                                                # Variable
   ;