import * as monaco from 'monaco-editor'
import 'isomorphic-fetch'

const editor = monaco.editor.create(document.getElementById('editor'), {
  value: [
    'def oh_hai():',
    '\treturn "Hello Skylark!"',
    '',
    'print(oh_hai())'
  ].join('\n'),
  language: 'python',
  minimap: false
})

const output = document.getElementById('output')

const submit = document.getElementById('submit')
submit.addEventListener('pointerup', (e) => {
  output.innerText = ''

  return fetch('/exec?', {
    method: 'POST',
    body: editor.getValue()
  })
    .then(response => {
      output.setAttribute('class', response.ok ? 'result' : 'error')
      return response.text()
    }).then((text) => {
      output.innerText = text
    })
})
