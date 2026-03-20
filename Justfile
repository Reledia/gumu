set shell := ["fish", "-c"]
program_name := "gumu"

build:
  @go build -o {{program_name}}

run *args: build
  @go install .
  @{{program_name}} {{args}}

complete: build
  @./{{program_name}} completion -c fish | source

