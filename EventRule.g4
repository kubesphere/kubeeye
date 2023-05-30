grammar EventRule;

// Tokens
AND: 'and';
OR: 'or';
NOT: 'not' | '!';
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
LIKE: 'like';
NOTLIKE: 'not like';
REGEX: 'regex';
NOTREGEX: 'not regex';
EXISTS: 'exists';
NOTEXISTS: 'not exists';
COMMA: ',';
NUMBER: [ -]?[0-9]+('.'[0-9]+)?;
BOOLEAN: 'True'|'TRUE'|'true'|'False'|'FALSE'|'false';
STRING: '"' (ESC|.)*? '"';
//VAR: [a-zA-Z0-9_.-]+;
VAR: [a-zA-Z_][a-zA-Z0-9_]*('['('*'|[0-9]+|(([0-9]+)?':'([0-9]+)?))']')?('.'[a-zA-Z_][a-zA-Z0-9_]*('['('*'|[0-9]+|(([0-9]+)?':'([0-9]+)?))']')?('.')?)*;
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
   | VAR op=(EQU|NEQ) BOOLEAN                                           # BoolCompare
   | VAR op=(CONTAINS|NOTCONTAINS) (STRING|NUMBER)                      # ContainsOrNot
   | VAR op=(IN|NOTIN) '(' (NUMBER|STRING) (COMMA (NUMBER|STRING))* ')' # InOrNot
   | VAR op=(REGEX|NOTREGEX|LIKE|NOTLIKE) STRING                        # RegexOrNot
   | VAR op=(EXISTS|NOTEXISTS)                                          # ExistsOrNot
   | VAR                                                                # Variable
   | NOT VAR                                                            # NotVariable
   ;