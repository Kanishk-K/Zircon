window.addEventListener("load", function () {
  const player = this.document.querySelector("#playerScript");
  if (player === null) {
    logError("Player not found");
    return;
  } else {
    const partnerID = player.innerHTML.match(/"partnerId":(\d+)/)[1];
    const entryID = player.innerHTML.match(/entryId:"([^"]+)"/)[1];
    logSuccess(`Partner ID: ${partnerID}, Entry ID: ${entryID}`);
    const element = document.querySelector("#mediaContainer");

    getVideoInformation(partnerID, entryID).then((mediaInformation) => {
      console.log(mediaInformation);
      const downloadBanner = document.createElement("div");
      downloadBanner.classList.add("download-banner");
      if (mediaInformation.HD !== null) {
        const hd_download = document.createElement("a");
        hd_download.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32"  viewBox="0 0 256 256"><path d="M176,72H152a8,8,0,0,0-8,8v96a8,8,0,0,0,8,8h24a56,56,0,0,0,0-112Zm0,96H160V88h16a40,40,0,0,1,0,80Zm-64,8V136H56v40a8,8,0,0,1-16,0V80a8,8,0,0,1,16,0v40h56V80a8,8,0,0,1,16,0v96a8,8,0,0,1-16,0ZM24,48a8,8,0,0,1,8-8H224a8,8,0,0,1,0,16H32A8,8,0,0,1,24,48ZM232,208a8,8,0,0,1-8,8H32a8,8,0,0,1,0-16H224A8,8,0,0,1,232,208Z"></path></svg>`;
        hd_download.href = mediaInformation.HD.url;
        downloadBanner.appendChild(hd_download);
      }
      if (mediaInformation.SD !== null) {
        const sd_download = document.createElement("a");
        sd_download.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32"  viewBox="0 0 256 256"><path d="M144,72a8,8,0,0,0-8,8v96a8,8,0,0,0,8,8h24a56,56,0,0,0,0-112Zm64,56a40,40,0,0,1-40,40H152V88h16A40,40,0,0,1,208,128ZM24,48a8,8,0,0,1,8-8H224a8,8,0,0,1,0,16H32A8,8,0,0,1,24,48ZM232,208a8,8,0,0,1-8,8H32a8,8,0,0,1,0-16H224A8,8,0,0,1,232,208ZM104,152c0-9.48-8.61-13-26.88-18.26C61.37,129.2,41.78,123.55,41.78,104c0-18.24,16.43-32,38.22-32,15.72,0,29.18,7.3,35.12,19a8,8,0,1,1-14.27,7.22C97.64,91.93,89.65,88,80,88c-12.67,0-22.22,6.88-22.22,16,0,7,9,10.1,23.77,14.36C97.78,123,120,129.45,120,152c0,17.64-17.94,32-40,32s-40-14.36-40-32a8,8,0,0,1,16,0c0,8.67,11,16,24,16S104,160.67,104,152Z"></path></svg>`;
        sd_download.href = mediaInformation.SD.url;
        downloadBanner.appendChild(sd_download);
      }
      if (mediaInformation.transcript !== null) {
        const processButton = document.createElement("a");
        processButton.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32"  viewBox="0 0 256 256"><path d="M178.34,165.66,160,147.31V208a8,8,0,0,1-16,0V147.31l-18.34,18.35a8,8,0,0,1-11.32-11.32l32-32a8,8,0,0,1,11.32,0l32,32a8,8,0,0,1-11.32,11.32ZM160,40A88.08,88.08,0,0,0,81.29,88.68,64,64,0,1,0,72,216h40a8,8,0,0,0,0-16H72a48,48,0,0,1,0-96c1.1,0,2.2,0,3.29.12A88,88,0,0,0,72,128a8,8,0,0,0,16,0,72,72,0,1,1,100.8,66,8,8,0,0,0,3.2,15.34,7.9,7.9,0,0,0,3.2-.68A88,88,0,0,0,160,40Z"></path></svg>`;
        processButton.onclick = async function () {
          // Send a message to the background script to open the process page
          await chrome.runtime.sendMessage({
            type: "openProcessPage",
            mediaInformation: mediaInformation,
          });
        };
        downloadBanner.appendChild(processButton);
      }
      // add an event listener to the processButton to send to the server
      if (downloadBanner.children.length !== 0) {
        element.prepend(downloadBanner);
      }
    });
  }
});
