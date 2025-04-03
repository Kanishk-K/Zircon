class VideoSource {
  constructor(url, size, width, height) {
    this.url = url;
    this.size = size;
    this.width = width;
    this.height = height;
  }
}

async function getVideoInformation(partnerID, entryID) {
  // The purpose of this function is to return the following information
  // A video description containing
  //  - Title
  //  - Thumbnail URL
  // The URL of a video source best to give the user
  // The URL of a video source best to process (we really only care about the audio)

  const params = {
    // A multirequest essentially performs multiple operations in a single request [https://www.kaltura.com/api_v3/testmeDoc/general/multirequest.html]
    // This returns an array of size 4
    // 1. The session
    // 2. The flavor (video variation) data
    // 3. The actual video query data
    // 4. The attachment data (transcript)
    service: "multirequest",
    format: "1",
    ignoreNull: "1",
    "1:service": "session",
    "1:action": "startWidgetSession",
    "1:widgetId": `_${partnerID}`,
    "2:ks": "{1:result:ks}",
    "2:contextDataParams:flavorTags": "all",
    "2:service": "baseentry",
    "2:entryId": entryID,
    "2:action": "getContextData",
    "3:ks": "{1:result:ks}",
    "3:service": "baseentry",
    "3:action": "get",
    "3:entryId": entryID,
    "4:ks": "{1:result:ks}",
    "4:service": "attachment_attachmentasset",
    "4:action": "list",
    "4:filter:entryIdEqual": entryID,
  };
  const paramsAsString = new URLSearchParams(params).toString();
  const url = "https://cdnapi.kaltura.com/api_v3/index.php?" + paramsAsString;
  const response = await fetch(url);
  const data = await response.json();
  const title = data[2].name;
  const thumbnail = `${data[2].thumbnailUrl}/width/${data[2].width}`;
  const sources = convertToSource(data);
  const transcript = convertToTranscript(data);
  return {
    title: title,
    thumbnail: thumbnail,
    HD: sources.HD,
    SD: sources.SD,
    transcript: transcript,
    entryID: entryID,
  };
}

function convertToSource(data) {
  const partnerID = data[0].partnerId;
  const baseURL = `https://cdnapi.kaltura.com/p/${partnerID}/sp/${partnerID}00/playManifest`;
  const flavorData = data[1].flavorAssets;
  HDFlavor = null;
  SDFlavor = null;

  for (const asset of flavorData) {
    if (asset.status != 2 || asset.fileExt != "mp4") {
      // Status of 2 means that it's done and ready to be used
      // We also want to make sure that the file is an mp4 (potentially changed to process on server if needed)
      // If the data is not at this state, it is volatile and should not be used
      continue;
    }
    if (
      HDFlavor === null ||
      asset.width > HDFlavor.width ||
      asset.size < HDFlavor.size
    ) {
      HDFlavor = new VideoSource(
        `${baseURL}/entryId/${asset.entryId}/format/download/protocol/https/flavorParamIds/${asset.flavorParamsId}`,
        asset.size,
        asset.width,
        asset.height
      );
    }
    if (
      SDFlavor === null ||
      asset.width < SDFlavor.width ||
      asset.size < SDFlavor.size
    ) {
      SDFlavor = new VideoSource(
        `${baseURL}/entryId/${asset.entryId}/format/download/protocol/https/flavorParamIds/${asset.flavorParamsId}`,
        asset.size,
        asset.width,
        asset.height
      );
    }
  }
  return {
    HD: HDFlavor,
    SD: SDFlavor,
  };
}

function convertToTranscript(data) {
  const entryID = data[2].id;
  const attachments = data[3].objects;
  for (const attachment of attachments) {
    if (attachment.fileExt == "txt" && attachment.filename.includes(entryID)) {
      return `https://cdnapi.kaltura.com/api_v3/index.php/service/attachment_attachmentAsset/action/serve/attachmentAssetId/${attachment.id}`;
    }
  }
  return null;
}

// Takes in media information, creates a div with the download buttons and returns that element.
async function generateBanner(mediaInformation) {
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
    processButton.innerHTML = `<svg width="156" height="174" viewBox="0 0 156 174" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" clip-rule="evenodd" d="M146 137.74L155.942 132V93H146V137.74ZM146 81H155.942V42L146 36.2598V81ZM134 31V81H116.971V64.5L83.9999 45.4641V31H134ZM39.0288 81V64.5L71.9999 45.4641V31H21.9999V81H39.0288ZM39.0288 93H21.9999V143H71.9999V128.536L39.0288 109.5V93ZM83.9999 128.536V143H134V93H116.971V109.5L83.9999 128.536ZM116.105 155L83.9999 173.536V155H116.105ZM116.105 19H83.9999V0.464081L116.105 19ZM71.9999 19V0.464134L39.8949 19H71.9999ZM71.9999 173.536V155H39.8948L71.9999 173.536ZM9.99992 36.2599V81H0.0576782V42L9.99992 36.2599ZM9.99992 93V137.74L0.0576782 132V93H9.99992Z" fill="#C59F63"/></svg>`;
    processButton.onclick = async function () {
      // Send a message to the background script to open the process page
      await chrome.runtime.sendMessage({
        type: "openProcessPage",
        mediaInformation: mediaInformation,
      });
    };
    downloadBanner.appendChild(processButton);
  }
  return downloadBanner;
}
