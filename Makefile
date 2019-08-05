.PHONY: test

test:
	-@PATH=.:$$PATH bats test
