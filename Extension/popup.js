const REDIRECT_URL = chrome.identity.getRedirectURL("callback");
import { CLIENT_ID, AUTH_SERVERHOST, DOMAIN } from "./src/util/info.js";
import { getUserJWT, parseJwt } from "./src/util/jwt.js";

window.addEventListener("load", async (event) => {
  const AUTH_CONTENT = document.getElementById("logged-in-content");
  const UNAUTH_CONTENT = document.getElementById("login-content");
  const LOGIN_BTN = document.getElementById("login");

  function updateProfile() {
    chrome.storage.local.get(["name"], (data) => {
      document.getElementById("name").innerText = data.name;
    });
    chrome.storage.sync.get(["recentProcesses"], (result) => {
      const recentProcesses = result.recentProcesses || [];
      const recentProcessesList = document.getElementById("prev_analysis");
      recentProcesses.forEach((process) => {
        const process_card = document.createElement("div");
        process_card.className = "process-card";
        recentProcessesList.appendChild(process_card);

        const process_image = document.createElement("img");
        process_image.src = process.thumbnail;
        process_image.alt = "Lecture Image";
        process_card.appendChild(process_image);

        const process_details = document.createElement("div");
        process_details.className = "process-details";
        process_card.appendChild(process_details);

        const process_title = document.createElement("h3");
        process_title.innerText = process.title;
        process_details.appendChild(process_title);

        const process_actions = document.createElement("div");
        process_actions.className = "process-actions";
        process_details.appendChild(process_actions);

        const HD_BTN = document.createElement("a");
        HD_BTN.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32"  viewBox="0 0 256 256"><path d="M176,72H152a8,8,0,0,0-8,8v96a8,8,0,0,0,8,8h24a56,56,0,0,0,0-112Zm0,96H160V88h16a40,40,0,0,1,0,80Zm-64,8V136H56v40a8,8,0,0,1-16,0V80a8,8,0,0,1,16,0v40h56V80a8,8,0,0,1,16,0v96a8,8,0,0,1-16,0ZM24,48a8,8,0,0,1,8-8H224a8,8,0,0,1,0,16H32A8,8,0,0,1,24,48ZM232,208a8,8,0,0,1-8,8H32a8,8,0,0,1,0-16H224A8,8,0,0,1,232,208Z"></path></svg>`;
        HD_BTN.href = process.HD.url;
        HD_BTN.target = "_blank";
        process_actions.appendChild(HD_BTN);

        const SD_BTN = document.createElement("a");
        SD_BTN.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32"  viewBox="0 0 256 256"><path d="M144,72a8,8,0,0,0-8,8v96a8,8,0,0,0,8,8h24a56,56,0,0,0,0-112Zm64,56a40,40,0,0,1-40,40H152V88h16A40,40,0,0,1,208,128ZM24,48a8,8,0,0,1,8-8H224a8,8,0,0,1,0,16H32A8,8,0,0,1,24,48ZM232,208a8,8,0,0,1-8,8H32a8,8,0,0,1,0-16H224A8,8,0,0,1,232,208ZM104,152c0-9.48-8.61-13-26.88-18.26C61.37,129.2,41.78,123.55,41.78,104c0-18.24,16.43-32,38.22-32,15.72,0,29.18,7.3,35.12,19a8,8,0,1,1-14.27,7.22C97.64,91.93,89.65,88,80,88c-12.67,0-22.22,6.88-22.22,16,0,7,9,10.1,23.77,14.36C97.78,123,120,129.45,120,152c0,17.64-17.94,32-40,32s-40-14.36-40-32a8,8,0,0,1,16,0c0,8.67,11,16,24,16S104,160.67,104,152Z"></path></svg>`;
        SD_BTN.href = process.SD.url;
        SD_BTN.target = "_blank";
        process_actions.appendChild(SD_BTN);

        const PROCESS_BTN = document.createElement("a");
        PROCESS_BTN.innerHTML = `<svg width="156" height="174" viewBox="0 0 156 174" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" clip-rule="evenodd" d="M146 137.74L155.942 132V93H146V137.74ZM146 81H155.942V42L146 36.2598V81ZM134 31V81H116.971V64.5L83.9999 45.4641V31H134ZM39.0288 81V64.5L71.9999 45.4641V31H21.9999V81H39.0288ZM39.0288 93H21.9999V143H71.9999V128.536L39.0288 109.5V93ZM83.9999 128.536V143H134V93H116.971V109.5L83.9999 128.536ZM116.105 155L83.9999 173.536V155H116.105ZM116.105 19H83.9999V0.464081L116.105 19ZM71.9999 19V0.464134L39.8949 19H71.9999ZM71.9999 173.536V155H39.8948L71.9999 173.536ZM9.99992 36.2599V81H0.0576782V42L9.99992 36.2599ZM9.99992 93V137.74L0.0576782 132V93H9.99992Z" fill="#C59F63"/></svg>`;
        PROCESS_BTN.onclick = async function () {
          // Send a message to the background script to open the process page
          await chrome.runtime.sendMessage({
            type: "openProcessPage",
            mediaInformation: process,
          });
        };
        process_actions.appendChild(PROCESS_BTN);
      });
    });
  }

  const token = await getUserJWT();
  if (token) {
    updateProfile();
    console.log(token);
    AUTH_CONTENT.classList.remove("hidden");
  } else {
    LOGIN_BTN.addEventListener("click", async (event) => {
      chrome.identity.launchWebAuthFlow(
        {
          url: `${AUTH_SERVERHOST}/login?client_id=${CLIENT_ID}&response_type=code&scope=email+openid+profile&redirect_uri=${encodeURIComponent(
            REDIRECT_URL
          )}`,
          interactive: true,
        },
        async (redirectUrl) => {
          const url = new URL(redirectUrl);
          const code = url.searchParams.get("code");
          fetch(`${AUTH_SERVERHOST}/oauth2/token`, {
            method: "POST",
            headers: {
              "Content-Type": "application/x-www-form-urlencoded",
            },
            body: new URLSearchParams({
              grant_type: "authorization_code",
              code: code,
              client_id: CLIENT_ID,
              redirect_uri: REDIRECT_URL,
            }),
          })
            .then((res) => res.json())
            .then(async (data) => {
              if (!data.id_token) {
                throw new Error("No id_token in response");
              }
              const id_info = parseJwt(data.id_token);
              await chrome.cookies.set({
                url: `https://${DOMAIN}`,
                domain: DOMAIN,
                name: "id_token",
                value: data.id_token,
                secure: true,
                httpOnly: true,
                expirationDate: id_info.exp - 60,
              });
              await chrome.storage.local.set({
                refresh_token: data.refresh_token,
                name: id_info.name,
                email: id_info.email,
              });
              console.log(id_info);
              updateProfile();
              UNAUTH_CONTENT.classList.add("hidden");
              AUTH_CONTENT.classList.remove("hidden");
            });
        }
      );
    });
    UNAUTH_CONTENT.classList.remove("hidden");
  }
});
