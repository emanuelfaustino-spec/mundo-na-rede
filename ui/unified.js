"use strict";

// One integrated navigation; existing controls are moved, never duplicated.
(function initUnifiedWorkspace(){
  const world=document.querySelector('[data-section="world"]');
  const addons=document.querySelector('[data-section="addons"]');
  const sourceActions=$("addon-folder").closest(".actions");
  if(!world || !addons || !sourceActions) throw new Error("Estrutura de navegação incompleta.");

  const engineGroup=document.createElement("details");
  engineGroup.id="engine-install";
  engineGroup.className="engine-install";
  const engineTitle=document.createElement("summary");
  engineTitle.textContent=t("Preparar motor do servidor");
  const engineControls=document.createElement("div");
  engineControls.className="actions";
  for(const id of ["official","mixed-install","manual-install","eula","world"]){
    engineControls.append($(id));
  }
  engineGroup.append(engineTitle,$("install-tip"),engineControls,$("engine-explainer"));
  world.insertBefore(engineGroup,world.children[1]||null);

  $("addons-controls").append($("addon-folder").closest("label"),sourceActions.querySelector('[data-upload="addon"]'));
  for(const id of ["addon-tip","extension-list","internal-packs","catalog"]){
    $("addons-content").append($(id));
  }
  $("console-content").append($("console-details"));
  $("console-details").open=true;

  // Destructive controls only appear when specifically expanded.
  const dangerButton=$("delete-server");
  const danger=document.createElement("details");
  danger.className="danger-zone";
  const summary=document.createElement("summary");
  summary.textContent="Outras ações";
  dangerButton.replaceWith(danger);
  danger.append(summary,dangerButton);

  const expert=localStorage.getItem("mundo-nexus-expert-mode")==="true";
  document.body.classList.toggle("expert-mode",expert);
  $("expert-toggle").checked=expert;
  $("expert-toggle").addEventListener("change",event=>{
    document.body.classList.toggle("expert-mode",event.target.checked);
    localStorage.setItem("mundo-nexus-expert-mode",String(event.target.checked));
  });
  for(const id of ["pregeneration-panel","internal-packs"])$(id).classList.add("expert-only");

  const tooltips={
    "start":"Inicia o mundo para receber jogadores.",
    "stop":"Salva e encerra o servidor.",
    "restart":"Reinicia o motor do mundo.",
    "backup":"Cria uma cópia manual que só você pode excluir.",
    "restore-open":"Recupera arquivos de um backup anterior.",
    "official":"Baixa o motor Bedrock oficial.",
    "manual-install":"Use um motor que você já baixou.",
    "world":"Importa um mundo existente.",
    "network-test":"Confere portas e rede neste PC.",
    "console-help":"Explica um erro conhecido em poucas palavras.",
    "addon-folder":"Escolha onde o complemento será instalado."
  };
  for(const [id,message] of Object.entries(tooltips)){
    if($(id))$(id).title=message;
  }
  for(const [name,message] of Object.entries({
    "port":"Número que seus amigos usam para entrar.",
    "memory":"Máximo de memória que o motor Java poderá usar.",
    "seed":"Código que define a geração do mundo novo.",
    "players":"Quantidade máxima de jogadores simultâneos."
  })){
    for(const field of document.querySelectorAll('[name="'+name+'"]'))field.title=message;
  }

  $("account-settings-button").addEventListener("click",()=> $("account-open").click());
  $("full-guide").addEventListener("click",()=> $("guide-open").click());
  $("network-share").addEventListener("click",()=> $("invite-open").click());
  $("network-settings").addEventListener("click",()=> $("world-edit").click());
  $("network-test").addEventListener("click",()=> verifyServer("network-result"));
  $("check-server").addEventListener("click",()=> verifyServer("preflight-result"));
  $("console-help").addEventListener("click",explainRecentError);
  $("copy-diagnostic").addEventListener("click",async()=>{
    if(!selected)return;
    const lines=["Mundo Nexus — diagnóstico",selected.engine,currentStatus.state||"desconhecido",
      "Porta "+selected.port,$("error-explanation").textContent];
    try {
      await navigator.clipboard.writeText(lines.join("\n"));
      notice("Diagnóstico copiado, sem dados da conta.");
    } catch {notice("Não foi possível copiar o diagnóstico.",true);}
  });

  // Only the most useful explanation in each guide item.
  const messages=[
    "Crie o servidor, prepare o motor e toque em Iniciar.",
    "Compartilhe IP e porta com quem está na mesma rede.",
    "Em Mundo, importe um .mcworld com o servidor parado.",
    "Abra Jogadores para permissões; Console para comandos.",
    "Importe arquivos compatíveis em Complementos, com o servidor parado.",
    "Esta função depende do motor Java e das configurações do servidor.",
    "Salve uma cópia em Backups. Para Drive, conecte sua conta.",
    "Use Verificar servidor no Painel e confira o Console."
  ];
  document.querySelectorAll("#guide details").forEach((item,index)=>{
    const p=item.querySelector("p");
    if(p && messages[index])p.textContent=t(messages[index]);
    if(index===0)item.open=false;
  });
})();

async function verifyServer(outputId) {
  const server=selected;
  if(!server)return null;
  const box=$(outputId);
  box.hidden=false;
  box.textContent="Verificando…";
  try {
    const result=await(await request("/"+server.id+"/preflight")).json();
    if(server!==selected)return null;
    box.replaceChildren();
    for(const check of result.checks){
      const item=document.createElement("p");
      item.className="check-"+check.level;
      item.textContent=(check.level==="ok"?"✓ ":check.level==="warn"?"! ":"× ")+check.text;
      if(check.help)item.title=check.help;
      box.append(item);
    }
    if(!result.ready)box.classList.add("has-errors");
    else box.classList.remove("has-errors");
    return result;
  }catch(error){
    box.textContent=error.message;
    return null;
  }
}

function explainRecentError() {
  const lines=(fullLogs||[]).slice(-120).join("\n").toLowerCase();
  const patterns=[
    [/address already in use|bind.*failed|port.*in use|porta em uso/,"Porta ocupada. Pare outro servidor ou altere a porta em Mundo."],
    [/outofmemory|out of memory|cannot allocate memory|java heap space/,"Pouca memória. Feche programas ou diminua a RAM do servidor."],
    [/eula.*false|agree to the eula/,"Aceite a EULA do Java em Mundo."],
    [/unsupportedclassversionerror|java.lang.unsupportedclassversion/,"A versão do Java não é compatível com o motor."],
    [/missing.*mod|requires.*fabric|dependency.*missing/,"Verifique as dependências do mod em Complementos."],
    [/failed to load|could not load|corrupt/,"Arquivo não carregou. Confira a versão e restaure um backup, se necessário."]
  ];
  const found=patterns.find(([pattern])=>pattern.test(lines));
  $("error-explanation").textContent=found?found[1]:"Nenhum erro conhecido encontrado. Confira as últimas mensagens do Console.";
}
