all: output_file_parser

output_file_parser:
	cd cmd/output_file_parser; go build output_file_parser.go

clean:
	@rm -f cmd/output_file_parser/output_file_parser