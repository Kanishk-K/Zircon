const SERVERHOST = "https://analysis.socialcoding.net";
chrome.storage.local.clear(); // Remove before deploying to prod, deletes auth information on each reload.

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
    chrome.storage.local.get("AUTH", async function (data) {
      if (data.AUTH !== undefined && data.AUTH.expiry > Date.now() / 1000) {
        // If the user is authenticated, send them to the processing page.
        const tab = await chrome.tabs.create({
          url: chrome.runtime.getURL("static/html/process.html"),
        });
        await onTabLoaded(tab.id);
        await chrome.tabs.sendMessage(tab.id, {
          action: "setData",
          data: request.mediaInformation,
          jwt: data.AUTH.token,
        });
      } else {
        // The user is not authenticated, make them log in before moving forward
        await chrome.tabs.create({
          url: chrome.runtime.getURL("/popup.html"),
        });
      }
    });
  },
};

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  const { type } = request;
  RuntimeMessages[type](request);
});

/* AUTH PROCESSING */

// Function injected into the page to extract JSON payload
function extractJSONPayload() {
  try {
    // Assuming the JSON payload is available in the body or a specific tag
    const jsonContent = document.body?.innerText || null;

    // Try parsing the content as JSON
    if (jsonContent) {
      return JSON.parse(jsonContent);
    }
  } catch (error) {
    console.error("Error parsing JSON:", error);
  }
  return null; // Return null if JSON parsing fails
}

chrome.tabs.onUpdated.addListener(async (tabId, changeInfo, tab) => {
  // Check if the tab finished loading and matches the base URL
  if (
    changeInfo.status === "complete" &&
    tab.url.startsWith(SERVERHOST + "/callback")
  ) {
    try {
      // Inject a script to extract the JSON payload
      const result = await chrome.scripting.executeScript({
        target: { tabId },
        func: extractJSONPayload,
      });

      // Extract the data returned by the content script
      const jsonPayload = result[0]?.result;

      if (jsonPayload) {
        chrome.storage.local.set({ AUTH: jsonPayload }, function () {});
      } else {
        console.error("Failed to extract JSON payload.");
      }

      // Close the tab
      chrome.tabs.remove(tabId);
    } catch (error) {
      console.error("Error handling callback page:", error);
    }
  }
});
