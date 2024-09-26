# Notes on code generation

1. Mechanics of code generation are easy with libs such as jennifer

1. But usefulness, correctness and flexibility of generated code depends on the inputs

1. Using an adhoc type model eases understanding while generating code and reduces extra data stored for the job

1. But an adhoc model requires to input som uch details in a custom format that makes it unusable

1. Alternatively, the custom model can be filled from a base source code using reflection

1. Reflection is quite insufficient, reflection types do not carry labels or named values such as receiver names, argument names.

1. Some gaps from reflection can be filled by convention and extra settings

1. Some gaps can be filled by making short names from type names to mimic what you would normally use.

1. Sometimes this is unfeasible: `id string` -> `s string` no way to guess that string is an `id`.

1. Other times it is ugly: `lib.Resource` -> `libR`

1. Reading code directly with annotations parsed as an AST (Abstract Syntax Tree) seems a better approach, it is also the usual way for generation tools.

1. Reading code allows to run without compiling the input code in, which reflection requires because the types must exist in the process running the generation.

1. Question 1 for AST? does it also replace the internal data model. Most probably, but a PoC would make this clearer.

1. Question 2 for AST? Should it be used to also generate the final code or is using a generator such as jennifer a simpler option?
