# Contributing 

Just a regular PR process should work in general to get a change into the architecture documentation. 
However, there are different changes that would require a bit of preparation that is specified in this document.  

## Contribute or modify a diagram

- We use [mermaid](https://mermaid-js.github.io/mermaid/#/) as tooling for diagrams as code.
- We are using [c4model](https://c4model.com/) for diagramming. [Mermaid support for C4](https://mermaid-js.github.io/mermaid/#/c4c) 
  is emerging so not super stable.
- Currently, in order to update a diagram we suggest to follow the following flow:
  - Grab the diagram source code from this repo.
  - Edit it in [mermaid](https://mermaid.live/)
  - Once happy with the diagram, download a svg image and check in into the code. reference the diagram as image.
  - Add or update the diagram source code back to have it for future reference.
