.PHONY: all
all:
ifndef GOOGLE_APPLICATION_CREDENTIALS
  $(error GOOGLE_APPLICATION_CREDENTIALS is undefined)
endif
	echo dog > dog.txt
	go run .
