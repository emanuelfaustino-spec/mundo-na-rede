"use strict";

// Engine downloads begin only after the user accepts the official terms.
async function engineLibraryRequest(path="",options={}){
  const response=await fetch("/api/engines"+path,{
    ...options,
    headers:{"X-Local-Key":key||"",...(options.headers||{})}
  });
  const data=await response.json().catch(()=>({}));
  if(!response.ok) throw Error(data.error||"Não foi possível consultar a biblioteca.");
  return data;
}

async function showEngineLibrary(){
  const dialog=$("engines-dialog");
  if(!dialog.open)dialog.showModal();
  await refreshEngineLibrary();
}

async function refreshEngineLibrary(){
  const box=$("engines-status");
  try{
    const state=await engineLibraryRequest();
    const lines=[
      state.bedrockReady?"✓ Bedrock pronto":"• Bedrock ainda não preparado",
      state.paperReady?"✓ Paper pronto":"• Paper ainda não preparado",
      state.javaReady?"✓ Java 25 disponível no PC":state.portableJavaSupported?"• Java 25 ainda não preparado":"! Instale Java 25 neste PC para usar Paper"
    ];
    box.replaceChildren();
    for(const message of lines){
      const p=document.createElement("p");p.textContent=message;box.append(p);
    }
    $("engines-prepare").disabled=state.bedrockReady && state.paperReady && (state.javaReady || !state.portableJavaSupported);
    $("engines-prepare").textContent=state.bedrockReady && state.paperReady && (state.javaReady || !state.portableJavaSupported)
      ?"Motores preparados":"Preparar Bedrock e Java";
    return state;
  }catch(error){box.textContent=error.message;return null}
}

$("engines-close").addEventListener("click",()=>$("engines-dialog").close());
$("create-engine-open").addEventListener("click",showEngineLibrary);

const openLibrary=document.createElement("button");
openLibrary.type="button";openLibrary.textContent="Biblioteca de motores";
openLibrary.id="open-engine-library";
openLibrary.addEventListener("click",showEngineLibrary);
$("engine-install").append(openLibrary);

$("engines-prepare").addEventListener("click",async()=>{
  const message="Você aceita os termos do Minecraft e do Eclipse Temurin e autoriza baixar os componentes oficiais neste PC? Os arquivos serão reutilizados pelos servidores.";
  if(!confirm(message))return;
  const button=$("engines-prepare");
  button.disabled=true;
  const result=$("engines-result");
  let failed=0;
  const current=await refreshEngineLibrary();
  const items=[["bedrock","Bedrock"],["paper","Paper"]];
  if(current?.portableJavaSupported)items.push(["java","Java 25"]);
  for(const [kind,name] of items){
    result.textContent="Preparando "+name+"…";
    try{
      await engineLibraryRequest("/prepare?kind="+kind,{
        method:"POST",headers:{"X-Accept-Minecraft-Terms":"yes"}
      });
    }catch(error){failed++;result.textContent=name+": "+error.message;await new Promise(r=>setTimeout(r,350));}
  }
  const status=await refreshEngineLibrary();
  result.textContent=failed ? "Alguns motores não foram preparados. Confira a conexão e tente novamente."
    :status?.javaReady?"Biblioteca pronta. Crie seu servidor.":"Motores preparados. Instale Java compatível para servidores Java.";
  button.disabled=!status || (status.bedrockReady && status.paperReady && (status.javaReady || !status.portableJavaSupported));
});

setTimeout(async()=>{
  const status=await refreshEngineLibrary();
  if(!status)return;
  const introKey="mundo-nexus-engine-library-intro";
  if((!status.bedrockReady || !status.paperReady || (!status.javaReady && status.portableJavaSupported)) && !localStorage.getItem(introKey)){
    localStorage.setItem(introKey,"true");
    if(!$("engines-dialog").open)$("engines-dialog").showModal();
  }
},250);
