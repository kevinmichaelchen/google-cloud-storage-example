.PHONY: all
all:
ifndef GOOGLE_APPLICATION_CREDENTIALS
  $(error GOOGLE_APPLICATION_CREDENTIALS is undefined)
endif
	go run .
