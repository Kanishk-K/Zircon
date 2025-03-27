import { AUTH_SERVERHOST, CLIENT_ID, DOMAIN } from "./info.js";
export function parseJwt(token) {
  var base64Url = token.split(".")[1];
  var base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
  var jsonPayload = decodeURIComponent(
    atob(base64)
      .split("")
      .map(function (c) {
        return "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2);
      })
      .join("")
  );

  return JSON.parse(jsonPayload);
}

export async function getUserJWT() {
  return new Promise((resolve, reject) => {
    chrome.cookies.get(
      {
        url: `https://${DOMAIN}`,
        name: "id_token",
      },
      (cookie) => {
        if (cookie) {
          resolve(cookie.value);
        } else {
          // Try to refresh the token, if that fails delete the refresh token and return null
          chrome.storage.local.get("refresh_token", async (data) => {
            if (data.refresh_token) {
              try {
                const response = await fetch(
                  `${AUTH_SERVERHOST}/oauth2/token`,
                  {
                    method: "POST",
                    headers: {
                      "Content-Type": "application/x-www-form-urlencoded",
                    },
                    body: new URLSearchParams({
                      grant_type: "refresh_token",
                      refresh_token: data.refresh_token,
                      client_id: CLIENT_ID,
                    }),
                  }
                );

                const tokenData = await response.json();

                if (response.ok) {
                  const id_info = parseJwt(tokenData.id_token);
                  await chrome.cookies.set({
                    url: `https://${DOMAIN}`,
                    domain: DOMAIN,
                    name: "id_token",
                    value: tokenData.id_token,
                    secure: true,
                    httpOnly: true,
                    expirationDate: id_info.exp - 60,
                  });
                  await chrome.storage.local.set({
                    refresh_token: tokenData.refresh_token,
                    name: id_info.name,
                    email: id_info.email,
                  });
                  resolve(tokenData.id_token);
                } else {
                  await chrome.storage.local.remove("refresh_token");
                  resolve(null);
                }
              } catch (error) {
                console.error("Error refreshing token:", error);
                await chrome.storage.local.remove("refresh_token");
                resolve(null);
              }
            } else {
              resolve(null);
            }
          });
        }
      }
    );
  });
}
