package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/bufbuild/protocompile"
)

// Message represents a protobuf message for the template
type Message struct {
	Name   string
	Fields []Field
}

// Field represents a field in a protobuf message
type Field struct {
	Name string
	Type string
}

func parseProtoFile(protoFilePath string) (msg []*Message, err error) {
	compiler := protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{},
	}
	compile, err2 := compiler.Compile(context.Background(), protoFilePath)
	if err2 != nil {
		return nil, err2
	}

	for _, file := range compile {
		messages := file.Messages()
		fmt.Println(messages)
		fields := messages.ByName("Sample").Fields()
		m := &Message{
			Name:   "Sample",
			Fields: nil,
		}
		msg = append(msg, m)
		for i := range fields.Len() {
			field := fields.Get(i)
			fmt.Printf("Field: %s, Type: %s\n", field.Name(), field.Kind().String())
			m.Fields = append(m.Fields, Field{
				Name: string(field.Name()),
				Type: field.Kind().String(),
			})
		}
	}
	return msg, nil
}
func main() {
	// Define and parse the command-line flag
	protoFilePath := flag.String("proto", "./sample.proto", "path to the .proto file")
	port := flag.String("port", ":8080", "port to run the server on")
	flag.Parse()

	// Parse the .proto file
	messages, err := parseProtoFile(*protoFilePath)
	if err != nil {
		log.Fatalf("Failed to parse .proto file: %v", err)
	}

	// Create HTTP handlers for each message
	for _, message := range messages {
		http.HandleFunc("/"+message.Name, func(w http.ResponseWriter, r *http.Request) {
			generateHTMLForm(w, message)
		})
	}

	http.HandleFunc("/submit", HandleFormSubmit)

	// Start the HTTP server
	fmt.Println("Starting server at :8080")
	log.Fatal(http.ListenAndServe(*port, nil))
}

func generateHTMLForm(w http.ResponseWriter, message *Message) {
	const tmplStr = `
<!DOCTYPE html>
<html>
<head>
    <title>Protobuf Message Form</title>
</head>
<body>
    <form action="/submit" method="post">
        {{range .Fields}}
        <label for="{{.Name}}">{{.Name}} ({{.Type}}):</label>
        <input type="text" id="{{.Name}}" name="{{.Name}}"><br>
        {{end}}
        <input type="submit" value="Submit">
    </form>
</body>
</html>
`
	tmpl, err := template.New("form").Parse(tmplStr)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	err = tmpl.Execute(w, message)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}
}

// HandleFormSubmit handles the form submission and builds a protobuf message
func HandleFormSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Here you would unmarshal the form data into your protobuf message
	// For example:
	// var msg YourProtobufMessageType
	// err := json.NewDecoder(r.Body).Decode(&msg)
	// ...

	// After you have the protobuf message, you can serialize it
	// serializedData, err := proto.Marshal(&msg)
	// ...

	// Respond with serialized protobuf data or handle it as needed
	// w.Header().Set("Content-Type", "application/x-protobuf")
	// w.Write(serializedData)
}
