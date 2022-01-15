'use strict';

const WASM_URL = 'wasm/main.wasm';

var wasm;

function init() {
  const go = new Go();
  if ('instantiateStreaming' in WebAssembly) {
    WebAssembly.instantiateStreaming(fetch(WASM_URL), go.importObject).then(function (obj) {
      wasm = obj.instance;
      go.run(wasm);
    })
  } else {
    fetch(WASM_URL).then(resp =>
      resp.arrayBuffer()
    ).then(bytes =>
      WebAssembly.instantiate(bytes, go.importObject).then(function (obj) {
        wasm = obj.instance;
        go.run(wasm);
      })
    )
  }
}

const ADD_URL = '/add';
function add(a,b) {
    const data = {
        "a": a,
        "b": b
    }
    fetch(ADD_URL,
        {
            method: 'POST',
            body: JSON.stringify(data),
            headers: {
                'Content-Type': 'application/json'
            },            
        }
    ).then(response => response.json())
    .then(respData => {        
         const r = decrypt(respData.result);
         document.getElementById("result").value = r;
        }
    );
  
}


init();