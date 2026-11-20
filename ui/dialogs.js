"use strict";

// Dialogs use HTML <dialog> rather than the WebView2 origin-labeled
// window.confirm/alert/prompt surfaces. Modal queuing prevents races.
(() => {
  const root = document.getElementById("nexus-dialog");
  const heading = document.getElementById("nexus-dialog-title");
  const messageElement = document.getElementById("nexus-dialog-message");
  const inputBox = document.getElementById("nexus-dialog-input");
  const cancel = document.getElementById("nexus-dialog-cancel");
  const accept = document.getElementById("nexus-dialog-accept");
  let sequence = Promise.resolve();
  function present(options) {
    return new Promise((resolve) => {
      const {type="confirm", title, message="", confirmText, cancelText="Cancelar", tone="default", value=""} = options;
      heading.textContent = title || ({alert:"Aviso",confirm:"Confirmar ação",prompt:"Informação"}[type] || "Mundo Nexus");
      messageElement.textContent = String(message);
      inputBox.hidden = type !== "prompt";
      inputBox.value = value;
      cancel.hidden = type === "alert";
      cancel.textContent = cancelText;
      accept.textContent = confirmText || (type === "alert" ? "Entendi" : "Confirmar");
      accept.classList.toggle("danger", tone === "danger");
      root.dataset.tone = tone;
      const finish = () => {
        root.removeEventListener("close", finish);
        resolve(root.returnValue === "accept"
          ? (type === "prompt" ? inputBox.value : type === "alert" ? undefined : true)
          : type === "prompt" ? null : type === "alert" ? undefined : false);
      };
      root.addEventListener("close", finish);
      root.showModal();
      if (type === "prompt") inputBox.focus();
      else (type === "alert" ? accept : cancel).focus();
    });
  }
  function enqueue(options) {
    const result = sequence.then(() => present(options));
    sequence = result.then(() => {}, () => {});
    return result;
  }
  cancel.addEventListener("click", () => root.close("cancel"));
  accept.addEventListener("click", () => root.close("accept"));
  root.addEventListener("cancel", (event) => {
    event.preventDefault();
    root.close("cancel");
  });
  root.addEventListener("keydown", (event) => {
    if (event.key === "Enter" && !event.isComposing && event.target === inputBox) {
      event.preventDefault();
      root.close("accept");
    }
  });
  window.nexusConfirm = (message, options={}) => enqueue({...options,type:"confirm",message});
  window.nexusAlert = (message, options={}) => enqueue({...options,type:"alert",message});
  window.nexusPrompt = (message, options={}) => enqueue({...options,type:"prompt",message});
})();
