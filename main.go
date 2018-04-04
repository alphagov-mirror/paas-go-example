package main

import (
    "encoding/json"
    "io/ioutil"
    "fmt"
    "log"
    "net/http"
    "os"
    "sort"
    "strings"

    "github.com/morhekil/mw"
)

type Record struct {
    Item []Country `json:"item"`
}

type Country struct {
    Name string `json:"name"`
}

var countries []Country

func init() {
    register := getCountryRegister()
    countries = parseCountries(register)
}

func getCountryRegister() []byte {
    response, err := http.Get("https://country.register.gov.uk/records.json?page-index=1&page-size=999")
    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }

    defer response.Body.Close()
    register, _ := ioutil.ReadAll(response.Body)

    return register
}

func parseCountries(register []byte) []Country {
    var records map[string]Record
    json.Unmarshal(register, &records)

    var countries []Country
    for _, record := range records {
        countries = append(countries, record.Item[0])
    }

    return countries;
}

func matchLetters(letters string) []Country {
    var matched = countries

    for _, letter := range letters {
        matched = matchLetter(matched, letter)
    }

    return matched
}

func matchLetter(countries []Country, letter rune) []Country {
    var matched []Country

    for _, country := range countries {
        name := strings.ToLower(country.Name)

        if strings.Contains(name, string(letter)) {
            matched = append(matched, country)
        }
    }

    return matched
}

func handlerRoot(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")

    fmt.Fprintf(w, `Examples:<ul>
        <li><a href="/?letters=uk">?letters=uk</a></li>
        <li><a href="/?letters=ab">?letters=ab</a></li>
        <li><a href="/?letters=z">?letters=z</a></li>
        <li><a href="/?letters=spi">?letters=spi</a></li>
        <li><a href="/">All countries</a></li>
    </ul>

    This is an example Go application that uses GOV.UK
    Registers to lookup countries that contain some given letters.

    <br/><br/>
    <a href="https://github.com/alphagov/paas-go-example">GitHub Repository</a>`)

    queries := r.URL.Query()
    letters := queries.Get("letters")
    matches := matchLetters(letters)

    sort.Slice(matches, func(i int, j int) bool {
        return matches[i].Name < matches[j].Name
    })

    fmt.Fprintf(w, "<h2>Matched countries:</h2> <p>%s</p>", matches)
}

func main() {
    port := os.Getenv("PORT")

    app := http.NewServeMux()
    app.HandleFunc("/", handlerRoot)

    http.ListenAndServe(":" + port,
        mw.Chaotic("/chaotic")(app),
    )
}
