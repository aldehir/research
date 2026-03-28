I really want to create a web app that helps me read research papers, primarily
PDFs for now.

Here is what I would like:

1. PDF rendering, probably using a well-known PDF js library
2. Ability to select text in the PDF
3. Ability to create chat sessions with an LLM. Perhaps for the chat session,
   include surrounding text and then the selected text to narrow down on what I
   want to ask.
4. Store PDFs on the server so I can access them from anywhere I can access the
   web app.

Here are some constraints:

1. Backend written in Go.
2. Frontend written in TypeScript with Vite and SvelteKit 5. No Server-Side
   rendering.
3. RED/GREEN TDD workflow: FAIL -> PASS -> REFACTOR
4. Support Anthropic's Messages API (Claude will be my LLM of choice).


Taking this into account, let's create a PRD.md. Feel free to ask me any
questions.
