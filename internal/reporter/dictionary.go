package reporter

import (
	"regexp"
	"strings"
)

type ErrorTranslation struct {
	Pattern   *regexp.Regexp
	RootCause string // Supports regex group replacements ($1, $2, dll)
	Solution  string // Supports regex group replacements
}

var dictionary = []ErrorTranslation{
	{
		Pattern:   regexp.MustCompile(`^undefined: (.*)$`),
		RootCause: "Identifier '$1' is undefined.",
		Solution:  "Declare '$1' before using it, or check for typos.",
	},
	{
		Pattern:   regexp.MustCompile(`^unused variable: (.*)$`),
		RootCause: "Variable '$1' is declared but never used.",
		Solution:  "Remove '$1' if it is not needed, or use it in an expression.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use (.*) \(type (.*)\) as type (.*) in (.*)$`),
		RootCause: "Type mismatch: cannot assign '$1' (type $2) to type $3 in $4.",
		Solution:  "Ensure the types match or perform an explicit type conversion.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use (.*) \(variable of type (.*)\) as (.*) value in (.*)$`),
		RootCause: "Type mismatch: cannot use '$1' (type $2) as type $3 in $4.",
		Solution:  "Ensure the types match or perform an explicit type conversion.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use (.*) \((.*)\) as (.*) value in (.*)$`),
		RootCause: "Type mismatch: cannot assign '$1' ($2) to type $3 in $4.",
		Solution:  "Ensure the types match or perform an explicit type conversion.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot convert (.*) \(type (.*)\) to type (.*)$`),
		RootCause: "Invalid conversion: cannot convert '$1' from type $2 to $3.",
		Solution:  "Check if the types are compatible for conversion.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) imported and not used$`),
		RootCause: "Cabinet '$1' is required but never used.",
		Solution:  "Remove the 'require' statement for '$1' or use it in your code.",
	},
	{
		Pattern:   regexp.MustCompile(`^expected 'package', found '(.*)'$`),
		RootCause: "Invalid file header: expected 'cabinet', found '$1'.",
		Solution:  "Declare 'cabinet <name>' at the very top of your file.",
	},
	{
		Pattern:   regexp.MustCompile(`^expected 'EOF', found (.*)$`),
		RootCause: "Unexpected token '$1' at the end of file.",
		Solution:  "Remove the trailing '$1' or fix the preceding syntax.",
	},
	{
		Pattern:   regexp.MustCompile(`^syntax error: unexpected (.*), expected (.*)$`),
		RootCause: "Syntax error: unexpected '$1', expected '$2'.",
		Solution:  "Fix the syntax to match the expected '$2'.",
	},
	{
		Pattern:   regexp.MustCompile(`^syntax error: unexpected (.*)$`),
		RootCause: "Syntax error: unexpected token '$1'.",
		Solution:  "Remove or correct the unexpected token.",
	},
	{
		Pattern:   regexp.MustCompile(`^too many arguments in call to (.*)$`),
		RootCause: "Too many arguments provided in call to '$1'.",
		Solution:  "Remove the extra arguments to match the function signature.",
	},
	{
		Pattern:   regexp.MustCompile(`^not enough arguments in call to (.*)$`),
		RootCause: "Not enough arguments provided in call to '$1'.",
		Solution:  "Provide the missing arguments required by the function signature.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid operation: (.*) \(mismatched types (.*) and (.*)\)$`),
		RootCause: "Invalid operation: $1 (mismatched types $2 and $3).",
		Solution:  "Ensure both sides of the operator have the same type.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot assign to (.*)$`),
		RootCause: "Cannot assign value to $1 (immutable or invalid left-hand side).",
		Solution:  "Ensure you are assigning to a mutable variable or valid memory location.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use interface (.*) in conversion \(contains specific type constraints or is comparable\)$`),
		RootCause: "Cannot use interface $1 in conversion (contains specific type constraints or is comparable).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use \~ outside of interface or type constraint \(use \^ for bitwise complement\)$`),
		RootCause: "Cannot use ~ outside of interface or type constraint (use ^ for bitwise complement).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid copy: arguments (.*) and (.*) have different element types (.*) and (.*)$`),
		RootCause: "Invalid copy: arguments $1 and $2 have different element types $3 and $4.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid method expression (.*)\.(.*) \(needs pointer receiver \(\*(.*)\)\.(.*)\)$`),
		RootCause: "Invalid method expression $1.$2 (needs pointer receiver (*$3).$4).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^impossible type switch case: (.*)\\n\\t(.*) cannot have dynamic type (.*) (.*)$`),
		RootCause: "Impossible type switch case: $1\n\t$2 cannot have dynamic type $3 $4.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid use of type alias (.*) in recursive type \(see go\.dev/issue/50729\)$`),
		RootCause: "Invalid use of type alias $1 in recursive type (see go.dev/issue/50729).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^implicit assignment to unexported field (.*) in struct literal of type (.*)$`),
		RootCause: "Implicit assignment to unexported field $1 in struct declaration of type $2.",
		Solution:  "Capitalize the first letter of the field '$1' in the struct definition to export it.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot assign to (.*) \(neither addressable nor a map index expression\)$`),
		RootCause: "Cannot assign to $1 (neither addressable nor a map index expression).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^impossible type assertion: (.*)\\n\\t(.*) does not implement (.*) (.*)$`),
		RootCause: "Impossible type assertion: $1\n\t$2 does not implement $3 $4.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot clear (.*): argument must be \(or constrained by\) map or slice$`),
		RootCause: "Cannot clear $1: argument must be (or constrained by) map or slice.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot handle more than (.*) union terms \(implementation limitation\)$`),
		RootCause: "Cannot handle more than $1 union terms (implementation limitation).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^package requires newer Go version (.*) \(application built with (.*)\)$`),
		RootCause: "Cabinet requires newer Go version $1 (application built with $2).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^mismatched types (.*) \(previous argument\) and (.*) \(type of (.*)\)$`),
		RootCause: "Mismatched types $1 (previous argument) and $2 (type of $3).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid use of ListExpr for index expression (.*) with (.*) indices$`),
		RootCause: "Invalid use of ListExpr for index expression $1 with $2 indices.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot slice (.*): (.*) and (.*) have different underlying types$`),
		RootCause: "Cannot slice $1: $2 and $3 have different underlying types.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^select case must be send or receive \(possibly with assignment\)$`),
		RootCause: "Select case must be send or receive (possibly with assignment).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid use of \[\.\.\.\] array \(outside a composite literal\)$`),
		RootCause: "Invalid use of [...] array (outside a declaration).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^assignment operation (.*) requires single\-valued expressions$`),
		RootCause: "Assignment operation $1 requires single-valued expressions.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^receiver declares (.*), but receiver base type declares (.*)$`),
		RootCause: "Receiver declares $1, but receiver base type declares $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^mixture of field:value and value elements in struct literal$`),
		RootCause: "Mixture of field:value and value elements in struct declaration.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^internal error: value of (.*) should be a boolean constant$`),
		RootCause: "Internal error: value of $1 should be a boolean constant.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot convert (.*) \(in (.*)\) to type (.*) \(in (.*)\)$`),
		RootCause: "Cannot convert $1 (in $2) to type $3 (in $4).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^embedded field type cannot be a pointer to an interface$`),
		RootCause: "Embedded field type cannot be a pointer to an interface.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^undefined array length (.*) or missing type constraint$`),
		RootCause: "Undefined array length $1 or missing type constraint.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use \~ outside of interface or type constraint$`),
		RootCause: "Cannot use ~ outside of interface or type constraint.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*)\.(.*) undefined \(type (.*) has no method (.*)\)$`),
		RootCause: "$1.$2 undefined (type $3 has no method $4).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid use of \~ \(underlying type of (.*) is (.*)\)$`),
		RootCause: "Invalid use of ~ (underlying type of $1 is $2).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) arguments for (.*) \(expected (.*), found (.*)\)$`),
		RootCause: "$1 arguments for $2 (expected $3, found $4).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot import package as init \- init must be a func$`),
		RootCause: "Cannot import cabinet as init - init must be a func.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^excessively long constant: (.*)\.\.\. \((.*) chars\)$`),
		RootCause: "Excessively long constant: $1... ($2 chars).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^It is illegal to define a label that is never used\.$`),
		RootCause: "It is illegal to define a label that is never used..",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use (.*) in union \((.*) embeds comparable\)$`),
		RootCause: "Cannot use $1 in union ($2 embeds comparable).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid cycle in declaration: (.*) refers to itself$`),
		RootCause: "Invalid cycle in declaration: $1 refers to itself.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot define new methods on instantiated type (.*)$`),
		RootCause: "Cannot define new methods on instantiated type $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^arguments have type (.*), expected floating\-point$`),
		RootCause: "Arguments have type $1, expected floating-point.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use (.*) in union \((.*) contains methods\)$`),
		RootCause: "Cannot use $1 in union ($2 contains methods).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) (.*)= (.*) \(mismatched types (.*) and (.*)\)$`),
		RootCause: "$1 $2= $3 (mismatched types $4 and $5).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^shifted operand (.*) \(type (.*)\) must be integer$`),
		RootCause: "Shifted operand $1 (type $2) must be integer.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^list contains both named and anonymous parameters$`),
		RootCause: "List contains both named and anonymous parameters.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot define new methods on non\-local type (.*)$`),
		RootCause: "Cannot define new methods on non-local type $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot switch on (.*) \((.*) is not comparable\)$`),
		RootCause: "Cannot switch on $1 ($2 is not comparable).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use FakeImportC and go115UsesCgo together$`),
		RootCause: "Cannot use FakeImportC and go115UsesCgo together.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use \.\.\. in call to non\-variadic (.*)$`),
		RootCause: "Cannot use ... in call to non-variadic $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^pointers of (.*) must have identical base types$`),
		RootCause: "Pointers of $1 must have identical base types.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^assignment mismatch: (.*) but (.*) returns (.*)$`),
		RootCause: "Assignment mismatch: $1 but $2 returns $3.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) expects (.*) or (.*) arguments; found (.*)$`),
		RootCause: "$1 expects $2 or $3 arguments; found $4.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^too many values in struct literal of type (.*)$`),
		RootCause: "Too many values in struct declaration of type $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^duplicate index (.*) in array or slice literal$`),
		RootCause: "Duplicate index $1 in array or slice literal.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^parameterized receiver contains nil parameters$`),
		RootCause: "Parameterized receiver contains nil parameters.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^break not in for, switch, or select statement$`),
		RootCause: "Break not in for, switch, or select statement.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^too few values in struct literal of type (.*)$`),
		RootCause: "Too few values in struct declaration of type $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid recursive type: (.*) refers to itself$`),
		RootCause: "Invalid recursive type: $1 refers to itself.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^multiple\-value (.*) in single\-value context$`),
		RootCause: "Multiple-value $1 in single-value context.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^argument has type (.*), expected complex type$`),
		RootCause: "Argument has type $1, expected complex type.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^embedded field type cannot be unsafe\.Pointer$`),
		RootCause: "Embedded field type cannot be unsafe.Pointer.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot convert (.*) to type (.*) \(in (.*)\)$`),
		RootCause: "Cannot convert $1 to type $2 (in $3).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot convert (.*) \(in (.*)\) to type (.*)$`),
		RootCause: "Cannot convert $1 (in $2) to type $3.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^field (.*) is embedded via a pointer in (.*)$`),
		RootCause: "Field $1 is embedded via a pointer in $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use iota outside constant declaration$`),
		RootCause: "Cannot use iota outside constant declaration.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^2nd and 3rd index required in 3\-index slice$`),
		RootCause: "2nd and 3rd index required in 3-index slice.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid case (.*) in switch on (.*) \((.*)\)$`),
		RootCause: "Invalid case $1 in switch on $2 ($3).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use (.*) as (.*) value in (.*): (.*)$`),
		RootCause: "Cannot use $1 as $2 value in $3: $4.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid composite literal(.*) type (.*)(.*)$`),
		RootCause: "Invalid declaration$1 type $2$3.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid use of \.\.\. in conversion to (.*)$`),
		RootCause: "Invalid use of ... in conversion to $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot slice (.*): no specific type in (.*)$`),
		RootCause: "Cannot slice $1: no specific type in $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^duplicate field name (.*) in struct literal$`),
		RootCause: "Duplicate field name $1 in struct declaration.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^initialization cycle: (.*) refers to itself$`),
		RootCause: "Initialization cycle: $1 refers to itself.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^methods must have a unique non\-blank name$`),
		RootCause: "Methods must have a unique non-blank name.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid use of \~ \((.*) is an interface\)$`),
		RootCause: "Invalid use of ~ ($1 is an interface).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^maps of (.*) must have identical key types$`),
		RootCause: "Maps of $1 must have identical key types.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^method (.*)\.(.*) already declared at (.*)$`),
		RootCause: "Method $1.$2 already declared at $3.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid use of (.*) in selector expression$`),
		RootCause: "Invalid use of $1 in selector expression.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot assign to struct field (.*) in map$`),
		RootCause: "Cannot assign to struct field $1 in map.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid field name (.*) in struct literal$`),
		RootCause: "Invalid field name $1 in struct declaration.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid use of \.\.\. with built\-in (.*)$`),
		RootCause: "Invalid use of ... with built-in $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^too many arguments in conversion to (.*)$`),
		RootCause: "Too many arguments in conversion to $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use \.\.\. with (.*)\-valued (.*)$`),
		RootCause: "Cannot use ... with $1-valued $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^expression in (.*) must be function call$`),
		RootCause: "Expression in $1 must be function call.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^embedded field type cannot be a pointer$`),
		RootCause: "Embedded field type cannot be a pointer.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot call pointer method (.*) on (.*)$`),
		RootCause: "Cannot call pointer method $1 on $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*)(.*)(.*) \(non\-numeric type (.*)\)$`),
		RootCause: "$1$2$3 (non-numeric type $4).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^index (.*) is out of bounds \(>= (.*)\)$`),
		RootCause: "Index $1 is out of bounds (>= $2).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^non\-boolean condition in for statement$`),
		RootCause: "Non-boolean condition in for statement.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot close receive\-only channel (.*)$`),
		RootCause: "Cannot close receive-only channel $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) \(mismatched types (.*) and (.*)\)$`),
		RootCause: "$1 (mismatched types $2 and $3).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) \(mismatched types (.*) and (.*)\)$`),
		RootCause: "$1 (mismatched types $2 and $3).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^name (.*) not exported by package (.*)$`),
		RootCause: "Name $1 not exported by cabinet $2.",
		Solution:  "Capitalize the first letter of '$1' to export it from the cabinet.",
	},
	{
		Pattern:   regexp.MustCompile(`^non\-boolean condition in if statement$`),
		RootCause: "Non-boolean condition in if statement.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot assign (.*) to (.*) \(in (.*)\)$`),
		RootCause: "Cannot assign $1 to $2 (in $3).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^missing argument in conversion to (.*)$`),
		RootCause: "Missing argument in conversion to $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot convert (.*) to type (.*): (.*)$`),
		RootCause: "Cannot convert $1 to type $2: $3.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot assign (.*) \(in (.*)\) to (.*)$`),
		RootCause: "Cannot assign $1 (in $2) to $3.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^got (.*) type arguments but want (.*)$`),
		RootCause: "Got $1 type arguments but want $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^use of \.\(type\) outside type switch$`),
		RootCause: "Use of .(type) outside type switch.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use (.*) as (.*) value in (.*)$`),
		RootCause: "Cannot use $1 as $2 value in $3.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot slice unaddressable value (.*)$`),
		RootCause: "Cannot slice unaddressable value $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^shifted operand (.*) must be integer$`),
		RootCause: "Shifted operand $1 must be integer.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^constant result is not representable$`),
		RootCause: "Constant result is not representable.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid receiver type (.*) \((.*)\)$`),
		RootCause: "Invalid receiver type $1 ($2).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^package (.*); expected package (.*)$`),
		RootCause: "Cabinet $1; expected cabinet $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid else branch in if statement$`),
		RootCause: "Invalid else branch in if statement.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^multiple defaults \(first at (.*)\)$`),
		RootCause: "Multiple defaults (first at $1).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^index (.*) must be integer constant$`),
		RootCause: "Index $1 must be integer constant.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^illegal cycle in method declaration$`),
		RootCause: "Illegal cycle in method declaration.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^use of package (.*) not in selector$`),
		RootCause: "Use of cabinet $1 not in selector.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot declare main \- must be func$`),
		RootCause: "Cannot declare main - must be func.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^index (.*) out of bounds \[0:(.*)\]$`),
		RootCause: "Index $1 out of bounds [0:$2].",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^file requires newer Go version (.*)$`),
		RootCause: "File requires newer Go version $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot declare init \- must be func$`),
		RootCause: "Cannot declare init - must be func.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^no new variables on left side of :=$`),
		RootCause: "No new variables on left side of :=.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^array length (.*) must be constant$`),
		RootCause: "Array length $1 must be constant.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^method (.*)\.(.*) already declared$`),
		RootCause: "Method $1.$2 already declared.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^assignment mismatch: (.*) but (.*)$`),
		RootCause: "Assignment mismatch: $1 but $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot assign (.*) to (.*) in (.*)$`),
		RootCause: "Cannot assign $1 to $2 in $3.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid slice indices: (.*) < (.*)$`),
		RootCause: "Invalid slice indices: $1 < $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^missing type in composite literal$`),
		RootCause: "Missing type in declaration.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^non\-name (.*) on left side of :=$`),
		RootCause: "Non-name $1 on left side of :=.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) is not a selector expression$`),
		RootCause: "$1 is not a selector expression.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^undefined: (.*) \(but have (.*)\)$`),
		RootCause: "Undefined: $1 (but have $2).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^duplicate key (.*) in map literal$`),
		RootCause: "Duplicate key $1 in map literal.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^operator (.*) not defined on (.*)$`),
		RootCause: "Operator $1 not defined on $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot convert (.*) to type (.*)$`),
		RootCause: "Cannot convert $1 to type $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid embedded field type (.*)$`),
		RootCause: "Invalid embedded field type $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^incorrect expression switch case$`),
		RootCause: "Incorrect expression switch case.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^shift count (.*) must be integer$`),
		RootCause: "Shift count $1 must be integer.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) repeated on left side of :=$`),
		RootCause: "$1 repeated on left side of :=.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot close non\-channel (.*)$`),
		RootCause: "Cannot close non-channel $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) is not a method signature$`),
		RootCause: "$1 is not a method signature.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot receive from (.*): (.*)$`),
		RootCause: "Cannot receive from $1: $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^could not import (.*) \((.*)\)$`),
		RootCause: "Could not import $1 ($2).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^unexpected list of expressions$`),
		RootCause: "Unexpected list of expressions.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use comparable in union$`),
		RootCause: "Cannot use comparable in union.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) is not a boolean constant$`),
		RootCause: "$1 is not a boolean constant.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^unknown channel direction (.*)$`),
		RootCause: "Unknown channel direction $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) (.*) must not be negative$`),
		RootCause: "$1 $2 must not be negative.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid map key type (.*)(.*)$`),
		RootCause: "Invalid map key type $1$2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^continue not in for statement$`),
		RootCause: "Continue not in for statement.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid function literal (.*)$`),
		RootCause: "Invalid function literal $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot use _ as value or type$`),
		RootCause: "Cannot use _ as value or type.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot receive from (.*) (.*)$`),
		RootCause: "Cannot receive from $1 $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*)\.(.*) undefined \((.*)\)$`),
		RootCause: "$1.$2 undefined ($3).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) has no single field (.*)$`),
		RootCause: "$1 has no single field $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^ambiguous selector (.*)\.(.*)$`),
		RootCause: "Ambiguous selector $1.$2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^constant (.*) overflows (.*)$`),
		RootCause: "Constant $1 overflows $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid import path \((.*)\)$`),
		RootCause: "Invalid import path ($1).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^unknown syntax\.Decl node %T$`),
		RootCause: "Unknown syntax.Decl node %T.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^length and capacity swapped$`),
		RootCause: "Length and capacity swapped.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) requires (.*) or later$`),
		RootCause: "$1 requires $2 or later.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^branch statement: (.*) (.*)$`),
		RootCause: "Branch statement: $1 $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) \((.*)\) is not a type$`),
		RootCause: "$1 ($2) is not a type.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid continue label (.*)$`),
		RootCause: "Invalid continue label $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid lhs in range clause$`),
		RootCause: "Invalid lhs in range clause.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot take address of (.*)$`),
		RootCause: "Cannot take address of $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^missing init expr for (.*)$`),
		RootCause: "Missing init expr for $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^missing key in map literal$`),
		RootCause: "Missing key in map literal.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid receiver type (.*)$`),
		RootCause: "Invalid receiver type $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^use of untyped nil in (.*)$`),
		RootCause: "Use of untyped nil in $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid constant type (.*)$`),
		RootCause: "Invalid constant type $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^incorrect type switch case$`),
		RootCause: "Incorrect type switch case.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) (.*) must be integer$`),
		RootCause: "$1 $2 must be integer.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid array length (.*)$`),
		RootCause: "Invalid array length $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid use of AssertExpr$`),
		RootCause: "Invalid use of AssertExpr.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot send to (.*): (.*)$`),
		RootCause: "Cannot send to $1: $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^negative shift count (.*)$`),
		RootCause: "Negative shift count $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^unknown channel direction$`),
		RootCause: "Unknown channel direction.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^malformed constant: (.*)$`),
		RootCause: "Malformed constant: $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid shift count (.*)$`),
		RootCause: "Invalid shift count $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid break label (.*)$`),
		RootCause: "Invalid break label $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) is not an interface$`),
		RootCause: "$1 is not an interface.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^incorrect tag syntax: %q$`),
		RootCause: "Incorrect tag syntax: %q.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot send to (.*) (.*)$`),
		RootCause: "Cannot send to $1 $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^3\-index slice of string$`),
		RootCause: "3-index slice of string.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) for built\-in (.*)$`),
		RootCause: "$1 for built-in $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) (.*) (.*) \((.*)\)$`),
		RootCause: "$1 $2 $3 ($4).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) (.*) overflows int$`),
		RootCause: "$1 $2 overflows int.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^extra init expr at (.*)$`),
		RootCause: "Extra init expr at $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^branch statement: (.*)$`),
		RootCause: "Branch statement: $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot make (.*): (.*)$`),
		RootCause: "Cannot make $1: $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid package name _$`),
		RootCause: "Invalid cabinet name _.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) cannot be ordered$`),
		RootCause: "$1 cannot be ordered.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot call (.*): (.*)$`),
		RootCause: "Cannot call $1: $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^missing index for (.*)$`),
		RootCause: "Missing index for $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) is a method value$`),
		RootCause: "$1 is a method value.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) is not a pointer$`),
		RootCause: "$1 is not a pointer.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^unknown operator (.*)$`),
		RootCause: "Unknown operator $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^in call to (.*), (.*)$`),
		RootCause: "In call to $1, $2.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^no key:value expected$`),
		RootCause: "No key:value expected.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid use of \.\.\.$`),
		RootCause: "Invalid use of ....",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot assign to (.*)$`),
		RootCause: "Cannot assign to $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^constant (.*)overflow$`),
		RootCause: "Constant $1overflow.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^extra init expr (.*)$`),
		RootCause: "Extra init expr $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot indirect (.*)$`),
		RootCause: "Cannot indirect $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid append: (.*)$`),
		RootCause: "Invalid append: $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) is not constant$`),
		RootCause: "$1 is not constant.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^anonymous parameter$`),
		RootCause: "Anonymous parameter.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot declare (.*)$`),
		RootCause: "Cannot declare $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) is not a slice$`),
		RootCause: "$1 is not a slice.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^more than one index$`),
		RootCause: "More than one index.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) is not a type$`),
		RootCause: "$1 is not a type.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid copy: (.*)$`),
		RootCause: "Invalid copy: $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^invalid statement$`),
		RootCause: "Invalid statement.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) used as type$`),
		RootCause: "$1 used as type.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot slice (.*)$`),
		RootCause: "Cannot slice $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) is not a map$`),
		RootCause: "$1 is not a map.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^=> (.*) ➞ (.*)\\n$`),
		RootCause: "=> $1 ➞ $2\n.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cannot index (.*)$`),
		RootCause: "Cannot index $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*)(.*) \((.*)\)$`),
		RootCause: "$1$2 ($3).",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) is too large$`),
		RootCause: "$1 is too large.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^division by zero$`),
		RootCause: "Division by zero.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^undefined: (.*)$`),
		RootCause: "Undefined: $1.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^missing return$`),
		RootCause: "Missing return.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^struct literal$`),
		RootCause: "Struct declaration.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^empty union$`),
		RootCause: "Empty union.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^(.*) failed$`),
		RootCause: "$1 failed.",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
	{
		Pattern:   regexp.MustCompile(`^cycle to $`),
		RootCause: "Cycle to .",
		Solution:  "Review the syntax or logic and ensure it complies with 0xg rules.",
	},
}

func TranslateGoError(rawMsg string) (string, string, bool) {
	rawMsg = strings.TrimPrefix(rawMsg, "invalid operation: ")
	for _, rule := range dictionary {
		if rule.Pattern.MatchString(rawMsg) {
			rootCause := rule.Pattern.ReplaceAllString(rawMsg, rule.RootCause)
			solution := rule.Pattern.ReplaceAllString(rawMsg, rule.Solution)
			return rootCause, solution, true
		}
	}
	return rawMsg, "", false
}
