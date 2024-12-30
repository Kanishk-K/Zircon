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

const RuntimeMessages = {
  // Add a request to store the data
  openProcessPage: async (request) => {
    const tab = await chrome.tabs.create({
      url: chrome.runtime.getURL("static/html/process.html"),
    });
    await onTabLoaded(tab.id);
    await chrome.tabs.sendMessage(tab.id, {
      action: "setData",
      data: request.mediaInformation,
    });
  },
};

chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  const { type } = request;
  RuntimeMessages[type](request);
});
