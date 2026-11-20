"use strict";
const assert=require("node:assert/strict");
const fs=require("node:fs");
const path=require("node:path");
const root=path.join(__dirname,"..");
const html=fs.readFileSync(path.join(root,"ui","index.html"),"utf8");
const ids=[...html.matchAll(/\bid="([^"]+)"/g)].map(match=>match[1]);
const seen=new Set();
for(const id of ids){assert(!seen.has(id),"ID repetido no HTML: "+id);seen.add(id)}
const tabs=[...html.matchAll(/\bdata-tab="([^"]+)"/g)].map(m=>m[1]);
const sections=[...html.matchAll(/\bdata-section="([^"]+)"/g)].map(m=>m[1]);
const expected=["all","world","addons","people","console","backup","network","automation","account","help"];
assert.deepEqual([...tabs].sort(),[...expected].sort(),"Navegação incompleta");
assert.deepEqual([...sections].sort(),[...expected].sort(),"Conteúdo das abas incompleto");
for(const id of [
 "world-tabs","more-tabs","expert-toggle","create-more","start","stop","restart","address",
 "server-name","check-server","preflight-result","network-test","network-result",
 "addons-controls","addons-content","console-content","console-details",
 "player-add-form","backup-history","backup-history-list","backup","restore-open",
 "account-settings-button","account-open","guide-open","full-guide",
 "schedule-form","engine-tip","install-tip","engine-explainer",
 "official","mixed-install","manual-install","addon-folder","addon-tip","catalog","delete-server"
])assert(seen.has(id),"Elemento ausente: "+id);
for(const name of ["language.js","app.js","unified.js"])
 assert(html.includes('src="'+name+'"'),"Script ausente: "+name);
const app=fs.readFileSync(path.join(root,"ui","app.js"),"utf8");
const unified=fs.readFileSync(path.join(root,"ui","unified.js"),"utf8");
assert(app.includes('chooseTab("console")'),"Console não usa a aba independente");
assert(unified.includes("verifyServer"),"Verificação local ausente");
assert(!app.includes("data-section=\"details\""),"Referência antiga à aba Detalhes");
console.log("Interface verificada: 10 áreas, IDs únicos e controles essenciais.");
