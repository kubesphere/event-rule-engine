.PHONY:all visitor
antlr4=java -Xmx500M -cp "/usr/local/lib/antlr-4.7.1-complete.jar:$CLASSPATH" org.antlr.v4.Tool
all:visitor test
visitor:
	$(antlr4) -Xlog -Dlanguage=Go -no-listener -visitor -o visitor/parser EventRule.g4
test:
	go build  -o eventruleengine examples/test.go
