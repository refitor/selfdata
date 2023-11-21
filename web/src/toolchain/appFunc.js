// import videojs from 'video.js'

export function getRemoteAvatar(id) {
  return id
  // return `https://loremflickr.com/70/70/people?lock=${id}`
}

export function isAndroid() {
  const userAgent = navigator.userAgent || ''
  const appVersion = navigator.appVersion || ''
  const vendor = navigator.vendor || ''
  let ua = userAgent + ' ' + appVersion + ' ' + vendor
  ua = ua.toLowerCase()

  let reg = 'android'
  if (ua.indexOf('android') >= 0) {
    reg = /\bandroid[ /-]?([0-9.x]+)?/
  } else if (ua.indexOf('adr') >= 0) {
    if (ua.indexOf('mqqbrowser') >= 0) {
      reg = /\badr[ ]\(linux; u; ([0-9.]+)?/
    } else {
      reg = /\badr(?:[ ]([0-9.]+))?/
    }
  }
  return new RegExp(reg).test(ua)
}

export function ReSizePic(ThisPic, RePicWidth) {
  const TrueWidth = ThisPic.width
  const TrueHeight = ThisPic.height
  const Multiple = TrueWidth / RePicWidth
  ThisPic.width = RePicWidth
  ThisPic.height = TrueHeight / Multiple
}

export function IsVideo(url) {
  if (IsImage(url) || GetAudioType(url) !== '') {
    return false
  }
  return true
}

export function IsImage(url) {
  if (url === '' || url === undefined) {
    return false
  }
  if (url.indexOf('.jpg') > -1 || url.indexOf('.JPG') > -1 || url.indexOf('.jpeg') > -1) {
    return true
  }
  if (url.indexOf('.png') > -1) {
    return true
  }
  if (url.indexOf('.gif') > -1) {
    return true
  }
  return false
}

export function GetAudioType(url) {
  const isQQ = url.indexOf('qq') > -1
  const isKuwo = url.indexOf('kuwo') > -1
  const isKugou = url.indexOf('kugou') > -1
  const isMusic = url.indexOf('music') > -1
  if (isQQ || isKuwo || isKugou || isMusic) {
    return 'audio/mp3'
  }
  if (url.indexOf('.mp3') > -1) {
    return 'audio/mp3'
  }
  if (url.indexOf('.ogg') > -1) {
    return 'audio/ogg'
  }
  if (url.indexOf('.wma') > -1) {
    return 'audio/wma'
  }
  return ''
}

// export function getMimetype(src) {
//   alert(videojs.getMimetype(src))
//   return 'video/mp4'
// }

// url解析函数
// ?id=111&name=567  => {id:111,name:567}
export function URLParse(params){
  let obj = {};
  let reg = /[?&][^?&]+=[^?&%]+/g;
  let url = params;
  let arr = url.match(reg);
  console.log('match arr by URLParse url : ' + url + ', arr: ' + arr)
  if (arr !== undefined && arr !== null) {
    arr.forEach((item) => {
      let tempArr = item.substring(1).split('=');
      let key = decodeURIComponent(tempArr[0]);
      let val = decodeURIComponent(tempArr[1]);
      obj[key] = val;
    })
  }
  console.log(obj)
  console.log('URLParse params: ' + params + ', obj: ' + obj)
  return obj;
}

//链接可点击
export function GetContentNodes(content) {
  const arrRet = []
  if (content.indexOf('http') === -1) {
    const ret = {}
    ret['Text'] = content
    arrRet.push(ret)
    return arrRet
  }

  let workContent = content
  const doSplitContent = function() {
    var reg = /(http:\/\/|https:\/\/)((\w|=|\?|\.|\/|&|-)+)/g;
    workContent = workContent.replace(reg, `<refitURL>$1$2<refitURL>`);
    console.log('workContent: ' + workContent)
    console.log(workContent.split(`<refitURL>`))

    const ret = {}
    let nodes = workContent.split(`<refitURL>`)
    if (nodes.length === -1) {
      ret['Text'] = workContent
      return ret
    }
    ret['Text'] = " " + nodes[0]
    ret['URL'] = nodes[1]
    nodes = nodes.slice(2)
    workContent = nodes.join("")
    return ret
  }

  while (workContent.length > 0 && workContent !== "") {
    arrRet.push(doSplitContent()) 
  }
  return arrRet
}