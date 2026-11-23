these are prompts which I gave to LLM

---

### 1
yo, I'm planning to build a calculator web app and we will use React for frontend and Golang for backend. 

hard instructions for you: 
- write readable, easy to understand, maintainable, idiomatic, easily extendable, easily testable code
- keep code simple, straightforward and minimal which fulfills all asked requirements
- dont overcomplicate things
- dont implement extra features
- dont do extra work apart from whatever is told
- DONT commit OR push anything, I'll do that myself

now, let's start this way: 
- initialize git
- create two folders in this directory - frontend/ for React, TypeScript and Vite and backend/ for server code with Golang
- then start with frontend first with these functionalities for calculator:
    - add a nice favicon for our app
    - calculator UI should be responsive, simple and minimal like real physical calculator
    - use this four column layout for keypad:
    ```
        AC  ⌫   (   )
        xʸ  √   %   ÷
        7   8   9   ×
        4   5   6   −
        1   2   3   +
        0   0   .   =
    ```
    - two zeros mean one double width button for zero
    - calculator display should look simple and minimal like real physical calculator
    - a icon only history button at the bottom of calculations that handles a small block for recent history showing upto 3 recent calculations, when history button is clicked then history is popped out and shown, by default it should be hidden

build only UI with these instructions and keep actual calculation logic for later

---

### 2
this looks fine
now, lets start on backend in backend/
- I want you to use Go 1.22+ and use only standard library, no external dependencies, use net/http with server timeouts

Also, add input validation and error handling on frontend side. 

And, this is how backend REST API should look like:
- POST /api/calculate; accept `application/json`; request body as single JSON object having expression string : `{"expression":"(200 + 10%) / 2"}`
- backend API returns result in JSON format
- for success, result looks like `{"result":110}`
- for errors, result looks like`{"error":{"code":"division_by_zero","message":"cant divide by zero"}}`
- server should validate inputs
- during inputs validation, reject unknown fields and trailing JSON
- put a size limit of 16KB on request body
- send proper error messages and http response codes in API response for - division by zero, invalid JSON OR expression syntax, oversized requests, unsupported media types, arithmetic domain OR range errors, unsupported routes OR methods and other similar failures
- logical guidelines: logic wise calculations should work similar to calculator in iphone

follow coding instructions throughly and stop after this part

---

### 3
these changes look fine, now, I want you to implement React calculator behavior and connect it to Go API and keep fetch client code separate.
once this is done then add tests for both sides using built in test runners and follow these instructions:
- give each test a short comment clarifying test intent
- cover JSON contract, methods, content types, malformed requests, request size boundaries, concurrency, all basic operator pairings, precedence, nested groups, percentages, roots, powers, final rounding and domain and range errors
- cover keypad edits, repeat equals, grouping, history, validation limits, retries, stale responses after AC and API success and error handling
- include an API response of zero
- mock fetch for client tests and keep tests independent of running servers

once this is done, write a concise root README.md for someone opening repository for first time, README should:
- include a small brief app overview
- local startup commands
- keyboard shortcuts
- small brief of calculator logical and functional approach
- expression only API with curl and success and error JSON
- test and production build commands
- precise test coverage summary

---

### 4
nice, we have both frontend and backend working with exhaustive tests, now, I want you to add a minimal Docker setup. 
Create a dockerfile to run both frontend and backend and frontend should communicate with backend and frontend URL should be exposed to test the app.
and once this is done then update README.md to include docker setup instructions.

then finally I want you to check:
- entire project against implemented behavior, make sure that README examples, test layout, favicon, Vite proxy, Docker config and ignore rules match actual files
- confirm that dependency lockfiles are included and generated artifacts are excluded
- run go test -race ./..., go vet ./..., npm test, npm run build and docker compose config
- check UI and API together if environment supports it
- report what passed and what remains unverified



