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
    // This returns an array of size 3
    // 1. The session
    // 2. The flavor (video variation) data
    // 3. The actual video query data
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
