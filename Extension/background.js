import { getUserJWT } from "./src/util/jwt.js";
// import { DOMAIN } from "./src/util/info.js";
// chrome.storage.local.clear();
// chrome.cookies.remove({
//   url: `https://${DOMAIN}`,
//   name: "id_token",
// });

// Checks if tab is loaded
function onTabLoaded(tabId) {
  return new Promise((resolve) => {
    chrome.tabs.onUpdated.addListener(function onUpdated(id, change) {
      if (id === tabId && change.status === "complete") {
        chrome.tabs.onUpdated.removeListener(onUpdated);
        resolve();
      }
    });
  });
}

// openProcessPage creates a new page, waits for it to load, then sends the content of the process to it.
const RuntimeMessages = {
  // Add a request to store the data
  openProcessPage: async (request) => {
    const token = await getUserJWT();
    if (token) {
      // If the user is authenticated, send them to the processing page.
      const tab = await chrome.tabs.create({
        url: chrome.runtime.getURL("static/html/process.html"),
      });
      await onTabLoaded(tab.id);
      await chrome.tabs.sendMessage(tab.id, {
        action: "setData",
        data: request.mediaInformation,
      });
    } else {
      // The user is not authenticated, make them log in before moving forward
      await chrome.tabs.create({
        url: chrome.runtime.getURL("/popup.html"),
      });
    }
  },
};

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  const { type } = request;
  RuntimeMessages[type](request);
});
