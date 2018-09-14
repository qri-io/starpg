import * as monaco from 'monaco-editor'
import 'isomorphic-fetch'

const editor = monaco.editor.create(document.getElementById('editor'), {
  value: [
    'def oh_hai():',
    '\treturn "Hello Skylark!"',
    '',
    'print(oh_hai())'
  ].join('\n'),
  language: 'python'
})

const output = document.getElementById('output')
const config = document.getElementById('config')
const secrets = document.getElementById('secrets')

const submit = document.getElementById('submit')
submit.addEventListener('pointerup', (e) => {
  const params = {
    'config': config.value,
    'secrets': config.value
  }

  
  const esc = encodeURIComponent
  const query = Object.keys(params)
  .map(k => esc(k) + '=' + esc(params[k]))
  .join('&')
  
  console.log(params, query)
  return fetch('/qri?' + query, {
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
