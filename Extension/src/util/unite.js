async function getChildVideos(partnerID, entryID) {
  queryParams = {
    1: {
      service: "session",
      action: "startWidgetSession",
      widgetId: `_${partnerID}`,
    },
    2: {
      service: "baseEntry",
      action: "list",
      filter: {
        objectType: "KalturaBaseEntryFilter",
        parentEntryIdEqual: `${entryID}`,
        typeIn: "1,7",
      },
      responseProfile: {
        type: 1,
        fields: "id",
      },
      ks: "{1:result:ks}",
    },
    apiVersion: "3.3.0",
    format: 1,
    ks: "",
    clientTag: "html5:v3.17.43",
    partnerId: partnerID,
  };

  const response = await fetch(
    "https://cdnapisec.kaltura.com/api_v3/service/multirequest",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(queryParams),
    }
  );
  const data = await response.json();
  return data[1]["objects"];
}

// Takes in media information, creates a div with the download buttons and returns that element.
async function generateUniteCard(mediaInformation) {
  const uniteCard = document.createElement("div");
  const downloadBanner = document.createElement("div");
  downloadBanner.classList.add("download-banner");
  if (mediaInformation.thumbnail !== null) {
    const mainVideoImage = document.createElement("img");
    mainVideoImage.src = mediaInformation.thumbnail;
    uniteCard.appendChild(mainVideoImage);
  }
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
  uniteCard.appendChild(downloadBanner);
  return uniteCard;
}
