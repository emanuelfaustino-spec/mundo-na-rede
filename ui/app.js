"use strict";

const $ = (id) => document.getElementById(id);
const key = location.hash.slice(1) || sessionStorage.getItem("local-key");
if (location.hash) {
  sessionStorage.setItem("local-key", key);
  history.replaceState(null, "", location.pathname);
}

const engines = {
  bedrock: {
    name: t("Bedrock oficial"),
    url: "https://www.minecraft.net/pt-br/download/server/bedrock",
    tip: t("Use o ZIP oficial do Bedrock Dedicated Server para este sistema."),
    folders: ["auto", "resource_packs", "behavior_packs"],
  },
  pnx: {
    name: "PowerNukkitX",
    url: "https://github.com/PowerNukkitX/PowerNukkitX/releases",
    tip: t("Use o JAR oficial do PowerNukkitX e Java 21 ou superior."),
    folders: ["resource_packs", "plugins"],
  },
  java: {
    name: t("Servidor Java"),
    url: "https://www.minecraft.net/pt-br/download/server",
    tip: t("Use um JAR executável compatível e instale a versão de Java exigida pelo motor."),
    folders: ["plugins", "mods"],
  },
  both: {
    name: t("Java e Bedrock"),
    url: "https://geysermc.org/wiki/geyser/setup/",
    tip: t("Instala Paper, Geyser, Floodgate e ViaVersion. Requer Java 21. O mundo usa as regras do Minecraft Java."),
    folders: ["plugins"],
  },
};
let host;
let selected;
let upload;
let lastLogs;
let fullLogs = [];
let currentStatus = {};
let coverURL;
let people;
const cpuHistory=[];
let pending = false;
let refreshing = false;
const form = $("create-form");

async function request(path, options = {}) {
  const response = await fetch("/api/servers" + path, {
    ...options,
    headers: { ...options.headers, "X-Local-Key": key || "" },
  });
  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(t(error.error) || t("Falha na operação."));
  }
  return response;
}
function notice(message, error = false) {
  $("notice").textContent = message;
  $("notice").classList.toggle("error", error);
  $("notice").hidden = !message;
}
function showCreate() {
  selected = undefined;
  $("create").hidden = false;
  $("panel").hidden = true;
  drawServers();
}
$("delete-server").addEventListener("click", () => {
  if (!selected || !confirm("Excluir o servidor ‘"+selected.name+"’ e os arquivos locais dele? Faça ou confirme seus backups antes. Esta ação não pode ser desfeita.")) return;
  const server=selected;
  perform(async () => {
    await request("/"+server.id+"/delete", {method:"POST"});
    host.servers=host.servers.filter((entry)=>entry.id!==server.id);
    showCreate();
    notice("Servidor excluído.");
  });
});
function drawServers() {
  $("servers").replaceChildren();
  for (const server of host?.servers || []) {
    const entry = document.createElement("button");
    entry.className = "server-entry";
    entry.textContent = server.name;
    entry.setAttribute("aria-current", String(server.id === selected?.id));
    entry.addEventListener("click", () => select(server));
    $("servers").append(entry);
  }
}
function select(server) {
  selected = server;
  currentStatus = {};
  cpuHistory.length=0;
  drawCPU();
  loadCover(server);
  $("players-online").replaceChildren();
  $("people-search").value="";
  people=null;
  chooseTab("all");
  drawServers();
  const engine = engines[server.engine];
  $("server-name").textContent = server.name;
  $("engine-name").textContent = engine.name;
  $("address").textContent = host.ips.length
    ? host.ips.map((ip) => server.engine === "both" ? formatAddress(server, ip) : ip + ":" + server.port).join(" · ")
    : t("Conecte este dispositivo à rede local");
  $("download-engine").href = engine.url;
  $("install-tip").textContent = engine.tip;
  $("world").hidden = server.engine !== "bedrock";
  $("eula").hidden = server.engine !== "java";
  $("official").hidden = server.engine !== "bedrock";
  $("access-edit").hidden=server.engine!=="bedrock";
  $("internal-packs").hidden=server.engine!=="bedrock";
  $("internal-packs").open=false;
  $("internal-list").replaceChildren();
  $("mixed-install").hidden = server.engine !== "both";
  $("manual-install").hidden = server.engine === "both";
  $("pause-info").textContent = t(["java", "both"].includes(server.engine)
    ? "Pausa após 15 minutos vazio em versões Java compatíveis."
    : "Pausa automática indisponível neste motor Bedrock.");
  $("config-upload").textContent = server.engine === "pnx" ? t("Importar pnx.yml") : t("Importar server.properties");
  $("addon-folder").replaceChildren(...engine.folders.map((folder) => new Option(t({auto:"Add-on Bedrock",resource_packs:"Pacote de recursos Bedrock",behavior_packs:"Pacote de comportamento Bedrock",plugins:"Plugin do servidor",mods:"Mod Java"}[folder]), folder)));
  updateAddonTip();
  $("catalog").hidden = !["java", "both"].includes(server.engine);
  $("catalog-kind").options[0].disabled = server.engine !== "java";
  $("catalog-kind").value = server.engine === "both" ? "plugin" : "mod";
  $("catalog-results").replaceChildren();
  drawPublicAddresses(server);
  $("extension-list").replaceChildren();
  refreshExtensions(server);
  $("create").hidden = true;
  $("panel").hidden = false;
  $("state").textContent = t("Consultando");
  $("logs").textContent = "";
  lastLogs = undefined;
  fullLogs = [];
  $("log-search").value = "";
  refresh();
  showSchedules(server).catch((error) => notice(error.message, true));
}
function formatAddress(server, ip) {
  return "Java " + ip + ":" + server.port + " / Bedrock " + ip + ":19132";
}
function updateAddonTip() {
  if (selected?.engine === "bedrock") {
    $("addon-tip").textContent = t("Importe um add-on Bedrock com o servidor parado. Os pacotes reconhecidos são ativados ao iniciar.");
    return;
  }
  $("addon-tip").textContent = ["plugins", "mods"].includes($("addon-folder").value)
    ? t("Selecione um JAR compatível. Alguns loaders exigem instalação específica no motor.")
    : t("Selecione um ZIP, .mcpack ou .mcaddon com manifest.json na raiz. Vincule o pacote ao mundo no formato do motor.");
}
function engineTip() {
  const engine = form.elements.engine.value;
  $("memory-label").hidden = engine === "bedrock";
  $("create-bedrock").hidden=engine!=="bedrock";
  form.elements.port.value = ["java","both"].includes(engine) ? "25565" : "19132";
  if (host?.availableMemory) {
    form.elements.memory.value = Math.min(4096, Math.max(512, Math.floor(host.availableMemory / 2 / 256) * 256));
  }
  $("engine-tip").textContent = engines[engine].tip;
}
function renderLogs() {
  const query = $("log-search").value.trim().toLocaleLowerCase();
  const logs = fullLogs.filter((line) => !query || line.toLocaleLowerCase().includes(query)).join("\n");
  const message = logs || t(query ? "Nenhuma linha encontrada." : "Os logs aparecem aqui.");
  if (message === lastLogs) return;
  const box = $("logs");
  const atEnd = box.scrollHeight - box.scrollTop - box.clientHeight < 28;
  box.textContent = message;
  if (atEnd) box.scrollTop = box.scrollHeight;
  lastLogs = message;
}
$("log-search").addEventListener("input", renderLogs);
async function refresh() {
  const server = selected;
  if (!server || refreshing) return;
  refreshing = true;
  try {
    const data = await (await request("/" + server.id + "/status")).json();
    if (selected !== server) return;
    currentStatus = data;
    $("state").dataset.state = data.state;
    drawHealth(server,data);
    $("access-edit").disabled = !["parado","falhou"].includes(data.state);
    $("state").textContent = !data.installed ? t("Motor pendente") : t(data.state);
    $("start").disabled = (!data.installed && server.engine!=="bedrock") || !["parado", "falhou"].includes(data.state);
    $("stop").disabled = !["iniciando", "online"].includes(data.state);
    $("restart").disabled = data.state !== "online";
    $("force").hidden = data.state !== "encerrando";
    $("uptime").textContent = data.seconds ? Math.floor(data.seconds / 60) + t(" min em execução") : "";
    fullLogs = data.logs || [];
    renderLogs();
  } catch (error) {
    notice(error.message, true);
  } finally {
    refreshing = false;
  }
}
async function perform(callback) {
  if (pending) return;
  pending = true;
  notice(t("Processando…"));
  try {
    await callback();
    notice(t("Operação concluída."));
    if (selected) refreshExtensions(selected);
  } catch (error) {
    notice(error.message, true);
  } finally {
    pending = false;
    await refresh();
  }
}
function action(name, body, query = "") {
  const server = selected;
  if (!server) return;
  return perform(() => request("/" + server.id + "/" + name + query, { method: "POST", body }));
}
form.addEventListener("submit", (event) => {
  event.preventDefault();
  perform(async () => {
    const data = Object.fromEntries(new FormData(form));
    for (const key of ["port", "memory", "players"]) data[key] = Number(data[key]);
    const server = await (await request("", { method: "POST", body: JSON.stringify(data) })).json();
    host.servers.push(server);
    select(server);
    const addon=$("initial-addon").files[0];
    if (server.engine==="bedrock" && addon) await request("/"+server.id+"/addon?"+new URLSearchParams({folder:"auto",name:addon.name}),{method:"POST",body:addon});
  });
});
$("new").addEventListener("click", showCreate);
form.elements.engine.addEventListener("change", engineTip);
$("addon-folder").addEventListener("change", updateAddonTip);
$("start").addEventListener("click", () => {
  if (selected?.engine==="bedrock" && (!currentStatus.installed || !currentStatus.access || currentStatus.accessWarning)) openPreparation(true);
  else action("start");
});
$("stop").addEventListener("click", () => action("stop"));
$("restart").addEventListener("click", () => action("restart"));
$("official").addEventListener("click", () => openPreparation(false));
$("access-edit").addEventListener("click", () => openPreparation(false));
$("mixed-install").addEventListener("click", () => {
  const server = selected;
  if (!server || !confirm(t("Você leu e aceita a EULA do Minecraft em https://www.minecraft.net/eula?"))) return;
  perform(() => request("/" + server.id + "/crossplay", {method: "POST", headers: {"X-Accept-BDS-Terms":"yes"}}));
});
$("eula").addEventListener("click", () => {
  if (confirm(t("Você leu e aceita a EULA do Minecraft em https://www.minecraft.net/eula?"))) action("accept-eula");
});
for (const [id, command] of [["kick", "kick"], ["ban", "ban"], ["operator", "op"]]) {
  $(id).addEventListener("click", () => {
    $("command").value = command + " ";
    $("command").focus();
  });
}
$("force").addEventListener("click", () => {
  if (confirm(t("Forçar a parada pode perder alterações recentes. Continuar?"))) action("force");
});
$("copy").addEventListener("click", async () => {
  if (!selected || !host.ips.length) return;
  try {
    await navigator.clipboard.writeText(selected.engine === "both" ? formatAddress(selected,host.ips[0]) : host.ips[0] + ":" + selected.port);
    notice(t("Endereço copiado."));
  } catch {
    notice(t("Selecione o endereço para copiar manualmente."), true);
  }
});
$("console-form").addEventListener("submit", (event) => {
  event.preventDefault();
  action("command", $("command").value);
});
for (const button of document.querySelectorAll(".shortcut")) {
  button.addEventListener("click", () => action("command", button.dataset.command));
}
$("allowlist").addEventListener("click", () => {
  const prefix = selected?.engine === "bedrock" ? "allowlist" : "whitelist";
  $("command").value = prefix + " add ";
  $("command").focus();
});
for (const button of document.querySelectorAll("[data-upload]")) {
  button.addEventListener("click", () => {
    upload = button.dataset.upload;
    if (upload === "restore" && !confirm(t("Restaurar os arquivos deste servidor? A pasta atual será preservada."))) return;
    const file = $("file");
    file.value = "";
    file.accept = {
      config: selected?.engine === "pnx" ? ".yml,.yaml" : ".properties",
      restore: ".zip",
      world: ".zip,.mcworld",
      install: selected?.engine === "bedrock" ? ".zip" : ".jar",
      addon: ["plugins", "mods"].includes($("addon-folder").value) ? ".jar" : ".zip,.mcpack,.mcaddon",
    }[upload];
    file.click();
  });
}
$("file").addEventListener("change", () => {
  const picked = $("file").files[0];
  if (!picked) return;
  const query = upload === "addon" ? "?" + new URLSearchParams({ folder: $("addon-folder").value, name: picked.name }) : "";
  action(upload, picked, query);
});
$("backup").addEventListener("click", () => {
  const server = selected;
  if (!server) return;
  perform(async () => {
    const response = await request("/" + server.id + "/backup");
    const url = URL.createObjectURL(await response.blob());
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = "mundo-na-rede-" + server.id + ".zip";
    anchor.click();
    setTimeout(() => URL.revokeObjectURL(url), 30000);
 if ($("backup-history").open) loadBackupHistory(server);
  });
});

async function showSchedules(server, data) {
  const endpoint = "/" + server.id + "/schedules";
  const items = await (await request(endpoint, data ? { method: "POST", body: JSON.stringify(data) } : {})).json();
  if (server !== selected) return;
  $("schedules").replaceChildren();
  for (const item of items) {
    const row = document.createElement("div");
    row.className = "task";
    const label = document.createElement("span");
    label.textContent = item.command + t(" · a cada ") + item.minutes + " min";
    const remove = document.createElement("button");
    remove.textContent = t("Remover");
    remove.addEventListener("click", () => perform(() => showSchedules(server, { id: item.id })));
    row.append(label, remove);
    $("schedules").append(row);
  }
}
$("schedule-form").addEventListener("submit", (event) => {
  event.preventDefault();
  if (!selected) return;
  const data = Object.fromEntries(new FormData(event.currentTarget));
  perform(() => showSchedules(selected, { command: data.command, minutes: Number(data.minutes) }));
});

(async function initialize() {
  try {
    host = await (await request("")).json();
    const platform = { windows: "Windows", android: "Android", linux: "Linux" }[host.os] || host.os;
    $("device").textContent = platform + " · " + host.cores + t(" núcleos") +
      (host.totalMemory ? " · " + (host.totalMemory / 1024).toFixed(1) + " GiB RAM" : "");
    if (host.os === "android") {
      form.elements.engine.querySelector('[value="bedrock"]').disabled = true;
      form.elements.engine.value = "pnx";
    }
    engineTip();
    form.elements.players.value = host.recommendedPlayers || 4;
    $("create-capacity").textContent = capacityHint();
    if (host.servers.length) select(host.servers[0]);
    else showCreate();
    setInterval(refresh, 1400);
  } catch (error) {
    notice(error.message, true);
  }
})();

for (const id of ["guide", "account", "config-dialog", "world-settings", "prepare-dialog", "restore-dialog","invite-dialog"]) {
  document.querySelector('[data-close="' + id + '"]').addEventListener("click", () => $(id).close());
}
$("guide-open").addEventListener("click", () => $("guide").showModal());

async function accountRequest(path = "", method = "GET", body) {
  const res = await fetch("/api/account" + path, { method, body, headers: { "X-Local-Key": key || "", ...(body ? {"Content-Type":"application/json"} : {}) } });
  if (!res.ok) {
    const error = await res.json().catch(() => ({}));
    throw new Error(t(error.error || "Falha na operação."));
  }
  return res.json();
}
async function showAccount() {
  const account = await accountRequest();
  const status = {
    pending: "Backup aguardando envio.",
    uploaded: "O último envio terminou.",
    upload_error: "O envio falhou. O aplicativo tentará novamente.",
    snapshot_error: "Não foi possível preparar o backup.",
  };
  $("account-state").textContent = t(!account.available ? "Google Drive indisponível nesta instalação." :
    account.connected ? status[account.status] || "Google Drive conectado." : "Nenhuma conta conectada.");
	$("account-setup").hidden = !account.needsSetup;
	if (account.needsSetup) {
	  $("account-state").textContent=t("Configure seu cliente OAuth para conectar o Google Drive.");
	  $("account-setup").open=true;
	}
  $("account-name").textContent = account.bridge ? "" : [account.name, account.email].filter(Boolean).join(" · ");
  $("account-bridge").hidden = !account.bridge;
  $("account-pair").hidden = !account.bridge || account.connected;
  if (!account.connected) $("cloud-backups").replaceChildren();
  $("account-connect").disabled = !account.available;
  $("account-connect").hidden = account.connected || account.bridge;
  $("account-disconnect").hidden = !account.connected;
}
async function updateAccount(callback) {
  try {
    await callback();
    await showAccount();
  } catch (error) {
    $("account-state").textContent = error.message;
  }
}
$("account-open").addEventListener("click", () => {
  $("account").showModal();
  updateAccount(loadCloudBackups);
});
$("account-connect").addEventListener("click", () => {
  const tab = window.open("about:blank", "_blank");
  if (tab) tab.opener = null;
  updateAccount(async () => {
    try {
      const auth = await accountRequest("/connect", "POST");
      if (tab) tab.location.replace(auth.url);
      else location.assign(auth.url);
    } catch (error) {
      tab?.close();
      throw error;
    }
  });
});
$("account-pair").addEventListener("click", () => updateAccount(async () => {
  await accountRequest("/pair", "POST");
  await loadCloudBackups();
}));
$("account-disconnect").addEventListener("click", () => {
  if (confirm(t("Desconectar este aparelho da conta Google?"))) updateAccount(() => accountRequest("/disconnect", "POST"));
});
$("account-setup-form").addEventListener("submit", (event) => {
  event.preventDefault();
  const tab=window.open("about:blank","_blank");
  if (tab) tab.opener=null;
  updateAccount(async () => {
    await accountRequest("/config", "POST", JSON.stringify({clientID:$("google-client-id").value,clientSecret:$("google-client-secret").value}));
    $("google-client-secret").value="";
    const auth=await accountRequest("/connect", "POST");
    if (tab) tab.location.replace(auth.url);
    else location.assign(auth.url);
  });
});
setInterval(() => {
  if ($("account").open) updateAccount(async () => {});
}, 3500);


let editingServer;
$("config-edit").addEventListener("click", async () => {
  if (!selected) return;
  editingServer = selected.id;
  $("config-text").value = "";
  $("config-text").disabled = true;
  $("config-status").textContent = t("Consultando");
  $("config-dialog").showModal();
  try {
    const config = await (await request("/" + editingServer + "/config")).json();
    $("config-heading").textContent = config.name;
    $("config-text").value = config.text;
    $("config-text").disabled = false;
    $("config-status").textContent = "";
  } catch (error) {
    $("config-status").textContent = error.message;
  }
});
$("config-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  if ($("config-text").disabled) return;
  const button = event.currentTarget.querySelector("button");
  button.disabled = true;
  try {
    await request("/" + editingServer + "/config", {method:"POST",body:$("config-text").value});
    $("config-dialog").close();
    notice(t("Operação concluída."));
  } catch (error) {
    $("config-status").textContent = error.message;
  } finally {
    button.disabled = false;
  }
});

function capacityHint() {
  return t("Sugestão para começar") + " · " + (host?.recommendedPlayers || 4) + " " + t("jogadores") + ".";
}
function drawPublicAddresses(server) {
  const box = $("public-addresses");
  box.replaceChildren();
  for (const [edition, address] of [["Java", server.publicJava], ["Bedrock", server.publicBedrock]]) {
    if (!address) continue;
    const row = document.createElement("div");
    row.className = "connection";
    const info = document.createElement("span");
    info.textContent = t("Internet") + " · " + edition + " · " + address;
    const copy = document.createElement("button");
    copy.textContent = t("Copiar");
    copy.addEventListener("click", async () => {
      try {
        await navigator.clipboard.writeText(address);
        notice(t("Endereço copiado."));
      } catch {
        notice(t("Selecione o endereço para copiar manualmente."), true);
      }
    });
    row.append(info, copy);
    box.append(row);
  }
  box.hidden = !box.childElementCount;
}
async function refreshExtensions(server) {
  try {
    const entries = await (await request("/" + server.id + "/extensions")).json();
    if (selected?.id !== server.id) return;
    const box = $("extension-list");
    const types = {resource_packs:"Pacote de recursos Bedrock",behavior_packs:"Pacote de comportamento Bedrock",plugins:"Plugin do servidor",mods:"Mod Java"};
    box.replaceChildren();
    if (!entries.length) {
      const message = document.createElement("p");
      message.textContent = t("0 add-ons instalados. Use Importar extensão para adicionar um pacote.");
      box.append(message);
      return;
    }
    const count=document.createElement("p");
    count.textContent=entries.length+" "+t("extensões instaladas");
    box.append(count);
    for (const entry of entries) {
      const row = document.createElement("div");
      row.className = "task";
      const label=document.createElement("span");
      label.textContent=t(types[entry.kind])+" · "+entry.name+(entry.version?" · "+entry.version:"");
      row.append(label);
      if (entry.id) {
        const remove=document.createElement("button");
        remove.textContent=t("Remover");
        remove.addEventListener("click",() => {
          if (confirm(t("Remover este add-on do mundo?"))) action("remove-addon",JSON.stringify({id:entry.id,kind:entry.kind}));
        });
        row.append(remove);
      }
      box.append(row);
    }
  } catch (error) {
    if (selected?.id === server.id) $("extension-list").textContent = error.message;
  }
}
const settingsForm = $("world-settings-form");
$("world-edit").addEventListener("click", () => {
  if (!selected) return;
  for (const field of ["name", "mode", "port", "players", "memory", "publicJava", "publicBedrock","seed","difficulty","preset"]) {
    settingsForm.elements[field].value = selected[field] ?? "";
  }
  $("world-memory").hidden = selected.engine === "bedrock";
  $("settings-bedrock").hidden=selected.engine!=="bedrock";
  $("address-java").hidden = !["java","both"].includes(selected.engine);
  $("address-bedrock").hidden = selected.engine === "java";
  $("world-capacity").textContent = capacityHint();
  $("world-settings-status").textContent = "";
  $("world-settings").showModal();
});
settingsForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!selected) return;
  const server = selected;
  const values = Object.fromEntries(new FormData(settingsForm));
  for (const key of ["port", "players", "memory"]) values[key] = Number(values[key]);
  const button = settingsForm.querySelector("button");
  button.disabled = true;
  try {
    const saved = await (await request("/" + server.id + "/world-settings", {method:"POST",body:JSON.stringify(values)})).json();
    const index = host.servers.findIndex((entry) => entry.id === saved.id);
    host.servers[index] = saved;
    $("world-settings").close();
    if (selected?.id === saved.id) select(saved);
    else drawServers();
    notice(t("Operação concluída."));
  } catch (error) {
    $("world-settings-status").textContent = error.message;
  } finally {
    button.disabled = false;
  }
});

$("catalog-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!selected) return;
  const server = selected;
  const kind = $("catalog-kind").value;
  const result = $("catalog-results");
  const button = event.currentTarget.querySelector("button");
  button.disabled = true;
  result.textContent = t("Consultando");
  try {
    const query = new URLSearchParams({kind, q: $("catalog-query").value});
    const hits = await (await request("/" + server.id + "/catalog?" + query)).json();
    if (selected !== server) return;
    result.replaceChildren();
    if (!hits.length) result.textContent = t("Nenhuma extensão encontrada.");
    for (const hit of hits) {
      if (!/^[A-Za-z0-9_-]+$/.test(hit.slug)) continue;
      const article = document.createElement("div");
      const link = document.createElement("a");
      link.href = "https://modrinth.com/" + kind + "/" + hit.slug;
      link.target = "_blank";
      link.rel = "noopener noreferrer";
      link.textContent = hit.title;
      const description = document.createElement("p");
      description.textContent = hit.description;
      article.append(link, description);
      result.append(article);
    }
  } catch (error) {
    if (selected === server) result.textContent = error.message;
  } finally {
    button.disabled = false;
  }
});

async function loadCloudBackups() {
  const account = await accountRequest();
  const box = $("cloud-backups");
  box.replaceChildren();
  if (!account.connected) return;
  const copies = await accountRequest("/backups");
  copies.sort((a, b) => Date.parse(b.modified) - Date.parse(a.modified));
  for (const copy of copies) {
    const row = document.createElement("div");
    row.className = "task";
    const label = document.createElement("span");
    const when = new Date(copy.modified);
    label.textContent = engines[copy.engine]?.name + " · " +
      (Number.isNaN(when.getTime()) ? copy.name : when.toLocaleString(document.documentElement.lang)) +
      " · " + Math.max(1, Math.round(copy.size / 1048576)) + " MiB";
    const restore = document.createElement("button");
    restore.textContent = t("Restaurar do Drive");
    restore.disabled = selected?.engine !== copy.engine;
    restore.addEventListener("click", () => {
      if (!confirm(t("Restaurar o backup no mundo selecionado? A pasta anterior será preservada."))) return;
      $("account").close();
      action("cloud-restore", JSON.stringify({id: copy.id}));
    });
    const remove = document.createElement("button");
    remove.textContent = t("Excluir backup");
    remove.addEventListener("click", () => {
      if (!confirm(t("Excluir esta cópia do Google Drive? Esta ação não pode ser desfeita."))) return;
      perform(async () => { await accountRequest("/backups/delete", {method:"POST",body:JSON.stringify({id:copy.id})}); await loadCloudBackups(); });
    });
    row.append(label, restore, remove);
    box.append(row);
  }
  if (!copies.length) box.textContent = t("Nenhum backup do aplicativo no Drive.");
}

function chooseTab(name) {
  for (const tab of document.querySelectorAll("[data-tab]")) {
    tab.setAttribute("aria-pressed", String(tab.dataset.tab === name));
  }
  if (name==="people" && selected) refreshPeople(selected);
  if (name==="details" && selected) refreshMetrics(selected);
  for (const card of document.querySelectorAll("[data-section]")) {
    card.hidden = card.dataset.section !== name;
  }
}
for (const tab of document.querySelectorAll("[data-tab]")) {
  tab.addEventListener("click", () => chooseTab(tab.dataset.tab));
}

function drawHealth(server, data) {
  const summary = $("server-summary");
  summary.replaceChildren();
  const values = [t("Porta") + " " + server.port];
  if (data.players !== undefined && server.engine === "bedrock" && data.state === "online") values.push(data.players + "/" + server.players + " " + t("jogadores"));
  if (data.version) values.push("BDS " + data.version);
  if (data.memoryTotal) values.push("RAM PC · " + (data.memoryUsed/1024).toFixed(1) + " / " + (data.memoryTotal/1024).toFixed(1) + " GiB");
  if (data.performance?.serverMemory) values.push(t("RAM do servidor")+" · "+Math.round(data.performance.serverMemory/1048576)+" MiB");
  if (data.performance?.serverCPU!==undefined) {
    values.push(t("CPU servidor")+" "+data.performance.serverCPU.toFixed(1)+"%");
    cpuHistory.push(data.performance.serverCPU);
    if (cpuHistory.length>60) cpuHistory.shift();
    drawCPU();
  } else if (data.state==="parado" || data.state==="falhou") {
    cpuHistory.length=0;
    drawCPU();
  }
  if (data.state === "online" && data.performance?.serverCPU !== undefined) {
    const cpu=data.performance.serverCPU;
    values.push(cpu >= 85 ? t("Estabilidade: sobrecarregado") : cpu >= 60 ? t("Estabilidade: atenção") : t("Estabilidade: ótimo"));
  }
  if (data.performance?.hostCPU!==undefined) values.push(t("CPU PC")+" "+data.performance.hostCPU.toFixed(1)+"%");
  for (const text of values) {
    const badge = document.createElement("span");
    badge.className = "pill";
    badge.textContent = text;
    summary.append(badge);
  }
  $("player-summary").textContent = data.state === "online" && data.players !== undefined && server.engine === "bedrock" ? data.players + "/" + server.players + " " + t("jogadores") : "";
  $("access-warning").hidden = !data.accessWarning || data.state !== "online";
  $("access-warning").textContent = t("A lista de permitidos está vazia. Pare o servidor e abra Quem pode entrar?");
}
function setOptions(select, values, blank) {
  select.replaceChildren(new Option(t(blank), ""), ...values.map((value) => new Option(value==="legacy"?t("Usar BDS de outro perfil"):value, value)));
}
let prepareStart = false;
function openPreparation(start) {
  if (!selected) return;
  prepareStart = start;
  $("access-mode").value = currentStatus.access || "lan";
  $("approved-players").value = (currentStatus.approved || []).join("\n");
  $("approved-label").hidden = $("access-mode").value !== "approved";
  $("approved-players").required = !$("approved-label").hidden;
  $("engine-prepare").hidden = Boolean(currentStatus.installed);
  $("prepare-eula").checked = false;
  $("prepare-eula").required = !currentStatus.installed;
  setOptions($("prepare-version"), [...(currentStatus.versions || []),...(currentStatus.legacyMotor?["legacy"]:[])], (currentStatus.versions || []).length ? "Usar versão instalada mais recente" : "Baixar BDS oficial");
  $("prepare-firewall").hidden = host?.os !== "windows";
  $("prepare-status").textContent = "";
  const showConsoleWarning=start && localStorage.getItem("mnr-hide-console-warning")!=="true";
  $("console-warning").hidden=!showConsoleWarning;
  $("console-warning-choice").hidden=!showConsoleWarning;
  $("hide-console-warning").checked=false;
  $("prepare-submit").textContent = t(start ? "Preparar e iniciar" : "Salvar preparação");
  $("prepare-dialog").showModal();
}
$("access-mode").addEventListener("change", () => {
  $("approved-label").hidden = $("access-mode").value !== "approved";
  $("approved-players").required = !$("approved-label").hidden;
});
async function withProgress(server, output, initial, operation) {
  output.textContent = t(initial);
  const timer = setInterval(async () => {
    try {
      const data = await (await request("/" + server.id + "/progress")).json();
      if (data.phase) output.textContent = t(data.phase);
    } catch {}
  }, 650);
  try {return await operation();}
  finally {clearInterval(timer);}
}
$("prepare-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const server = selected;
  const button = $("prepare-submit");
  if (!server || button.disabled) return;
  button.disabled = true;
  try {
    if (prepareStart && $("hide-console-warning").checked) localStorage.setItem("mnr-hide-console-warning","true");
    const path="/" + server.id;
    const players=$("approved-players").value.split(/\r?\n/).map((name) => name.trim()).filter(Boolean);
    await request(path+"/access",{method:"POST",body:JSON.stringify({mode:$("access-mode").value,players})});
    if (!$("engine-prepare").hidden) {
      const version=$("prepare-version").value;
      await withProgress(server,$("prepare-status"),"Verificando biblioteca",() =>
        request(path+"/official?"+new URLSearchParams({version}),{method:"POST",headers:{"X-Accept-BDS-Terms":"yes"}}));
    }
    if (prepareStart) {
      $("prepare-status").textContent=t("Iniciando servidor");
      await request(path+"/start",{method:"POST"});
    }
    $("prepare-dialog").close();
    notice(t(prepareStart ? "Iniciando servidor" : "Operação concluída."));
    await refresh();
  } catch (error) {$("prepare-status").textContent=error.message;}
  finally {button.disabled=false;}
});
let restoreSelection;
$("restore-open").addEventListener("click", () => {
  restoreSelection=undefined;
  $("restore-file").value="";
  $("restore-file").disabled=false;
  $("restore-confirm").disabled=true;
  $("restore-engine-label").hidden=true;
  $("restore-preview").replaceChildren();
  $("restore-status").textContent="";
  $("restore-dialog").showModal();
});
$("restore-file").addEventListener("change", () => stageRestoreFile($("restore-file").files[0]));
async function stageRestoreFile(file) {
  const server=selected;
  if (!file || !server) return;
  restoreSelection=undefined;
  $("restore-preview").replaceChildren();
  $("restore-engine-label").hidden=true;
  $("restore-confirm").disabled=true;
  $("restore-progress").hidden=false;
  $("restore-file").disabled=true;
  try {
    const response=await withProgress(server,$("restore-status"),"Recebendo backup",() =>
      request("/"+server.id+"/restore-preview",{method:"POST",body:file}));
    const info=await response.json();
    restoreSelection={server,info};
    const present=[info.world&&"Mundo",info.config&&"Configurações",info.engine&&"Motor Bedrock incluído"].filter(Boolean);
    $("restore-preview").textContent=present.map(t).join(" · ")+" · "+(info.size/1048576).toFixed(2)+" MiB · "+info.packs+" "+t("extensões instaladas")+(info.version?" · BDS "+info.version:"");
    if (!info.engine && server.engine==="bedrock") {
      const versions=currentStatus.versions || [];
      setOptions($("restore-version"),[...versions,...(currentStatus.legacyMotor?["legacy"]:[])],currentStatus.installed?"Manter motor deste perfil":"Preparar motor depois de restaurar");
      if (versions.includes(info.version)) $("restore-version").value=info.version;
      else if (!currentStatus.installed && versions.length) $("restore-version").value=versions[0];
      else if (!currentStatus.installed && currentStatus.legacyMotor) $("restore-version").value="legacy";
      $("restore-engine-label").hidden=false;
    }
    $("restore-status").textContent=t("Backup validado. Confira os arquivos antes de confirmar.");
    $("restore-confirm").disabled=false;
  } catch(error) {$("restore-status").textContent=error.message;}
  finally {$("restore-progress").hidden=true;$("restore-file").disabled=false;}
}
$("restore-confirm").addEventListener("click", async () => {
  if (!restoreSelection) return;
  const {server,info}=restoreSelection;
  $("restore-confirm").disabled=true;
  $("restore-file").disabled=true;
  $("restore-progress").hidden=false;
  const version=$("restore-engine-label").hidden?"":$("restore-version").value;
  try {
    await withProgress(server,$("restore-status"),"Restaurando arquivos",() =>
      request("/"+server.id+"/restore-apply",{method:"POST",body:JSON.stringify({token:info.token,version})}));
    restoreSelection=undefined;
    notice(t("Backup restaurado. A pasta anterior foi preservada."));
    $("restore-dialog").close();
    if (selected===server) {
      await refresh();
      refreshExtensions(server);
      if (server.engine==="bedrock" && !currentStatus.installed) openPreparation(false);
    }
  } catch(error) {$("restore-status").textContent=error.message;$("restore-confirm").disabled=false;}
  finally {$("restore-progress").hidden=true;$("restore-file").disabled=false;}
});
$("internal-packs").addEventListener("toggle", async (event) => {
  const server=selected;
  if (!event.currentTarget.open || !server) return;
  try {
    const rows=await (await request("/"+server.id+"/extensions?advanced=1")).json();
    if (selected!==server) return;
    const list=document.createElement("ul");
    for (const row of rows) {
      const item=document.createElement("li");
      item.textContent=row.name;
      list.append(item);
    }
    $("internal-list").replaceChildren(list);
  } catch(error) {$("internal-list").textContent=error.message;}
});

$("random-seed").addEventListener("click", () => {
  form.elements.seed.value=crypto.getRandomValues(new BigInt64Array(1))[0].toString();
});

async function loadCover(server) {
  if (coverURL) URL.revokeObjectURL(coverURL);
  $("world-cover").hidden=true;
  try {
    const result=await request("/"+server.id+"/cover");
    const blob=await result.blob();
    if (selected!==server) return;
    coverURL=URL.createObjectURL(blob);
    $("world-cover").src=coverURL;
    $("world-cover").hidden=false;
  } catch {}
}
$("world-cover-input").addEventListener("change", () => {
 const image=$("world-cover-input").files[0], server=selected;
 if (!image || !server) return;
 perform(async () => {
   await request("/"+server.id+"/cover",{method:"POST",body:image});
   await loadCover(server);
 });
});
let qrURL;
$("invite-open").addEventListener("click", async () => {
  const server=selected;
  if (!server) return;
  $("invite-address").textContent=$("address").textContent;
  $("invite-image").hidden=true;
  $("invite-dialog").showModal();
  if (qrURL) URL.revokeObjectURL(qrURL);
  if (!host?.ips?.length) return;
  try {
    const image=await (await request("/"+server.id+"/invite?"+new URLSearchParams({ip:host.ips[0]}))).blob();
    qrURL=URL.createObjectURL(image);
    $("invite-image").src=qrURL;
    $("invite-image").hidden=false;
  } catch {}
});

async function refreshPeople(server) {
  if (server.engine!=="bedrock") return;
  try {
    const data=await (await request("/"+server.id+"/players")).json();
    if (selected!==server) return;
    people=data;
    renderPeople();
  } catch {}
}
function renderPeople() {
  if (!people) return;
  const box=$("players-online");
  box.replaceChildren();
  const filter=$("people-search").value.trim().toLocaleLowerCase();
  const names=[...new Set([...people.online,...people.allowed,...people.banned,...Object.keys(people.recent).sort((a,b)=>Date.parse(people.recent[b])-Date.parse(people.recent[a]))])];
  for (const name of names) {
    if (!name.toLocaleLowerCase().includes(filter)) continue;
    const row=document.createElement("div");
    row.className="task";
    const label=document.createElement("span");
    const isBanned=people.banned.some((item)=>item.toLocaleLowerCase()===name.toLocaleLowerCase());
    label.textContent=name+(people.online.includes(name)?" · "+t("online"):"")+(people.allowed.includes(name)?" · "+t("Permitido"):"")+(isBanned?" · "+t("Banido"):"")+
      (people.recent[name]?" · "+new Date(people.recent[name]).toLocaleString(document.documentElement.lang):"");
    row.append(label);
    {
      const controls=document.createElement("div");
      controls.className="actions";
      const choices=isBanned?[["unban","Remover banimento"]]:[...(people.allowed.includes(name)?[["unallow","Remover da lista"]]:[["allow","Permitir"]]),["ban","Banir"]];
      if (currentStatus.state==="online" && people.online.includes(name)) choices.push(["kick","Expulsar"],["op","Dar permissão de operador"],["deop","Remover operador"]);
      for (const [cmd,title] of choices) {
        const button=document.createElement("button");
        button.textContent=t(title);
        button.addEventListener("click", async () => {
          if (cmd!=="allow" && !confirm(t(title)+" · "+name+"?")) return;
          await action("player-action",JSON.stringify({name,action:cmd}));
          setTimeout(() => {if (selected) refreshPeople(selected);},600);
        });
        controls.append(button);
      }
      row.append(controls);
    }
    box.append(row);
  }
}
$("people-search").addEventListener("input",renderPeople);
$("player-add-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!selected || selected.engine!=="bedrock") return;
  const name=$("player-add-name").value.trim();
  const action=$("player-add-action").value;
  if (!name) return;
  try {
    await actionRequest(name,action);
    $("player-add-name").value="";
  } catch(error) {notice(error.message,true);}
});
async function actionRequest(name, action) {
  if (["ban","unallow","unban","kick","deop"].includes(action) && !confirm(t({ban:"Banir",unallow:"Remover da lista",unban:"Remover banimento",kick:"Expulsar",deop:"Remover operador"}[action])+" · "+name+"?")) return;
  await action("player-action",JSON.stringify({name,action}));
  setTimeout(() => {if (selected) refreshPeople(selected);},300);
}
setInterval(() => {
  if(selected && document.querySelector('[data-tab="people"][aria-pressed="true"]')) refreshPeople(selected);
},5000);
async function refreshMetrics(server) {
  $("world-metrics").textContent="";
  try {
    const metrics=await (await request("/"+server.id+"/world-details")).json();
    if (selected!==server) return;
    const bits=[t("Tamanho do mundo")+" "+(metrics.bytes/1048576).toFixed(1)+" MiB"];
    for (const [key,title] of [["view-distance","Visão"],["tick-distance","Simulação"]]) {
      if (metrics.distances[key]) bits.push(t(title)+" "+metrics.distances[key]+" chunks");
    }
    $("world-metrics").textContent=bits.join(" · ");
    const performance=currentStatus.performance;
    if (currentStatus.state!=="online") $("performance-status").textContent=t("Inicie o servidor para acompanhar CPU e memória em tempo real.");
    else if (!performance || performance.serverCPU===undefined) $("performance-status").textContent=t("Coletando a primeira amostra de desempenho…");
    else $("performance-status").textContent=t("CPU do servidor")+" "+performance.serverCPU.toFixed(1)+"% · "+t("CPU PC")+" "+(performance.hostCPU ?? 0).toFixed(1)+"% · "+t("RAM do servidor")+" "+Math.round((performance.serverMemory ?? 0)/1048576)+" MiB";
  } catch {}
}

function drawPregenEstimate() {
  const radius=Math.max(100,Math.min(2000,Number($("pregenerate-radius").value)||0));
  const chunks=Math.ceil(Math.PI*radius*radius/256);
  const minutes=Math.max(1,Math.ceil(chunks/140));
  $("pregenerate-estimate").textContent=t("Estimativa")+": ~"+chunks.toLocaleString(document.documentElement.lang)+" chunks · "+t("pelo menos")+" ~"+minutes+" min. Espaço e tempo reais variam muito com seed, estruturas e add-ons.";
	$("pregenerate-radius-value").value=radius.toLocaleString(document.documentElement.lang)+" "+t("blocos");
}
$("pregenerate-radius").addEventListener("input",drawPregenEstimate);
drawPregenEstimate();
$("economy-mode").addEventListener("click", async () => {
  if (!selected || selected.engine!=="bedrock") return;
  if (!["parado","falhou"].includes(currentStatus.state)) { notice(t("Pare o servidor antes de aplicar o modo econômico."),true); return; }
  if (!confirm(t("Aplicar distância leve na próxima inicialização?"))) return;
  const values={...selected,preset:"light"};
  try {
    const saved=await (await request("/"+selected.id+"/world-settings",{method:"POST",body:JSON.stringify(values)})).json();
    host.servers[host.servers.findIndex((entry)=>entry.id===saved.id)]=saved;
    select(saved);
    notice(t("Modo econômico preparado para a próxima inicialização."));
  } catch(error) {notice(error.message,true);}
});

const storedTheme=localStorage.getItem("theme");
if (storedTheme) document.documentElement.dataset.theme=storedTheme;
$("theme").addEventListener("click", () => {
  const next=document.documentElement.dataset.theme==="light"?"dark":"light";
  document.documentElement.dataset.theme=next;
  localStorage.setItem("theme",next);
});
$("backup-history").addEventListener("toggle", () => {
  if (selected && $("backup-history").open) loadBackupHistory(selected);
});
async function loadBackupHistory(server) {
  const base="/"+server.id+"/backup-history";
  try {
    const history=await (await request(base)).json();
    if (selected!==server) return;
    const panel=$("backup-history-list");
    panel.replaceChildren();
    if (!history.length) panel.textContent=t("Nenhum backup local criado neste perfil.");
    for (const item of history) {
      const row=document.createElement("div");
      row.className="task";
      const label=document.createElement("span");
      label.textContent=new Date(item.date).toLocaleString(document.documentElement.lang)+" · "+(item.size/1048576).toFixed(1)+" MiB";
      const controls=document.createElement("div");
      controls.className="actions";
      const restore=document.createElement("button");
      restore.textContent=t("Restaurar backup");
      const remove=document.createElement("button");
      remove.textContent=t("Excluir backup");
      const route=base+"?"+new URLSearchParams({name:item.name});
      restore.addEventListener("click", async () => {
        $("restore-open").click();
        $("restore-status").textContent=t("Recebendo backup");
        try {await stageRestoreFile(await (await request(route)).blob());}
        catch(error){$("restore-status").textContent=error.message;}
      });
      remove.addEventListener("click", () => {
        if (!confirm(t("Excluir essa cópia local?"))) return;
        perform(async () => {await request(route,{method:"POST"});await loadBackupHistory(server);});
      });
      const upload=document.createElement("button");
      upload.textContent=t("Enviar ao Drive");
      upload.addEventListener("click",() => perform(() => request(route+"&upload=1",{method:"POST"})));
      controls.append(restore,upload,remove);
      row.append(label,controls);
      panel.append(row);
    }
  } catch(error) {$("backup-history-list").textContent=error.message;}
}

function drawCPU() {
  const canvas=$("cpu-chart");
  canvas.hidden=cpuHistory.length===0;
  const graphic=canvas.getContext("2d");
  graphic.clearRect(0,0,canvas.width,canvas.height);
  graphic.fillStyle="#4fa986";
  const bar=canvas.width/60;
  cpuHistory.forEach((value,index) => {
    const height=Math.max(2,canvas.height*Math.min(100,value)/100);
    graphic.fillRect(index*bar,canvas.height-height,bar-1,height);
  });
}

// Decorative star field stays deterministic per tab and avoids a canvas render loop.
(() => {
  const field=document.createElement("div");
  field.id="space-field";
  field.setAttribute("aria-hidden","true");
  let seed=0x9e3779b9;
  const random=() => ((seed=(seed*1664525+1013904223)>>>0)/4294967296);
  for(let i=0;i<150;i++){
    const star=document.createElement("i");
    const size=(random()*2.1+.45).toFixed(2)+"px";
    star.style.setProperty("--x",(random()*100).toFixed(2)+"%");
    star.style.setProperty("--y",(random()*100).toFixed(2)+"%");
    star.style.setProperty("--s",size);
    star.style.setProperty("--d",(random()*26+14).toFixed(1)+"s");
    star.style.setProperty("--delay",(-random()*35).toFixed(1)+"s");
    star.style.setProperty("--drift",(random()*90-45).toFixed(0)+"px");
    star.style.opacity=(random()*.68+.22).toFixed(2);
    field.append(star);
  }
  document.body.prepend(field);
})();

async function checkForUpdate() {
  try {
    const response=await fetch("/api/update",{headers:{"X-Local-Key":key||""}});
    const update=await response.json();
    if(!update.available) return;
    const automatic=localStorage.getItem("mnr-auto-update")==="true";
    const apply=async()=>{
      const result=await fetch("/api/update/apply",{method:"POST",headers:{"X-Local-Key":key||""}});
      if(!result.ok) throw new Error((await result.json().catch(()=>({}))).error||"Não foi possível baixar a atualização.");
      notice("Atualizando o Mundo na Rede… o aplicativo será reaberto.");
    };
    if(automatic){await apply();return;}
    if(confirm(`Atualização ${update.version} disponível. Deseja atualizar agora?`)){await apply();return;}
    if(confirm("Deseja ativar atualizações automáticas para as próximas versões?")) localStorage.setItem("mnr-auto-update","true");
  } catch(error) { console.debug("Verificação de atualização indisponível",error); }
}
setTimeout(checkForUpdate,1200);
