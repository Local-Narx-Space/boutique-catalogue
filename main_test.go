package main
import("encoding/json";"net/http/httptest";"testing")
func TestProducts(t *testing.T){w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET","/products",nil));var p []Product;if json.Unmarshal(w.Body.Bytes(),&p)!=nil||len(p)!=4||w.Code!=200{t.Fatal(w.Body.String())};for _,v:=range p{if v.Price<=0{t.Fatal("invalid price")}}}
func TestMissing(t *testing.T){w:=httptest.NewRecorder();handler().ServeHTTP(w,httptest.NewRequest("GET","/products/missing",nil));if w.Code!=404{t.Fatal(w.Code)}}
