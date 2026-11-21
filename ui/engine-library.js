"use strict";

// Shared engine library. Only missing components are downloaded.
async function engineLibraryRequest(path="",options={}){
  const response=await fetch("/api/engines"+path,{
    ...options,
    headers:{"X-Local-Key":key||"",...(options.headers||{})}
  });
  const data=await response.json().catch(()=>({}));
  if(!response.ok) throw Error(data.error||"Não foi possível consultar a biblioteca.");
  return data;
}
const engineNames={bedrock:"Bedrock",paper:"Paper",java:"Java 25"};
let engineLibraryBusy=false;

function missingEngines(status){
  if(!status)return [];
  return [
    ...(!status.bedrockReady?["bedrock"]:[]),
    ...(!status.paperReady?["paper"]:[]),
    ...(status.portableJavaSupported && !status.javaReady?["java"]:[])
  ];
}

async function showEngineLibrary(){
  const dialog=$("engines-dialog");
  if(!dialog.open)dialog.showModal();
  await refreshEngineLibrary();
}

async function refreshEngineLibrary(){
  const box=$("engines-status");
  const button=$("engines-prepare");
  try{
    const status=await engineLibraryRequest();
    const lines=[
      status.bedrockReady?"✓ Bedrock pronto":"• Bedrock ainda não preparado",
      status.paperReady?"✓ Paper pronto":"• Paper ainda não preparado",
      status.javaReady?"✓ Java disponível":status.portableJavaSupported?"• Java 25 ainda não preparado":"• Instale Java 25 para usar Paper"
    ];
    box.replaceChildren();
    for(const message of lines){
      const row=document.createElement("p");
      row.textContent=message;
      box.append(row);
    }
    const missing=missingEngines(status);
    button.disabled=engineLibraryBusy||missing.length===0;
    button.textContent=missing.length===0?"Biblioteca pronta":missing.length===1
      ?"Preparar "+engineNames[missing[0]]:"Preparar motores";
    return status;
  }catch(error){
    box.textContent=error.message;
    button.disabled=true;
    return null;
  }
}

$("engines-close").addEventListener("click",()=>$("engines-dialog").close());
$("create-engine-open").addEventListener("click",showEngineLibrary);
const openLibrary=document.createElement("button");
openLibrary.type="button";
openLibrary.textContent="Biblioteca de motores";
openLibrary.id="open-engine-library";
openLibrary.addEventListener("click",showEngineLibrary);
$("engine-install").append(openLibrary);

$("engines-prepare").addEventListener("click",async()=>{
  if(engineLibraryBusy)return;
  const state=await refreshEngineLibrary();
  if(!state)return;
  const missing=missingEngines(state);
  if(!missing.length)return;
  const selectedNames=missing.map(kind=>engineNames[kind]).join(", ");
  const accepted=await nexusConfirm(
    "Você aceita os termos do Minecraft e do Eclipse Temurin e autoriza preparar "+selectedNames+" neste PC? Os arquivos serão reutilizados pelos servidores.",
    {title:"Preparar motores",confirmText:"Aceitar e preparar",cancelText:"Cancelar"}
  );
  if(!accepted)return;
  engineLibraryBusy=true;
  const button=$("engines-prepare");
  const result=$("engines-result");
  button.disabled=true;
  result.textContent="";
  const problems=[];
  try{
    for(const kind of missing){
      result.textContent="Preparando "+engineNames[kind]+"…";
      result.setAttribute("aria-busy","true");
      try{
        await engineLibraryRequest("/prepare?kind="+encodeURIComponent(kind),{
          method:"POST",
          headers:{"X-Accept-Minecraft-Terms":"yes"}
        });
      }catch(error){
        problems.push(engineNames[kind]+": "+error.message);
      }
      await refreshEngineLibrary();
    }
  }finally{
    engineLibraryBusy=false;
    result.removeAttribute("aria-busy");
    const finalStatus=await refreshEngineLibrary();
    if(problems.length){
      result.textContent=problems.join("\n");
      result.classList.add("error");
    }else{
      result.classList.remove("error");
      result.textContent=missingEngines(finalStatus).length===0
        ?"Biblioteca pronta.":"Alguns motores ainda não estão preparados.";
    }
  }
});
