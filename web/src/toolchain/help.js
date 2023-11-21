function webResponse(response, failedCallback, successedCallback) {
    if (response.data['Error'] !== '' && response.data['Error'] !== null && response.data['Error'] !== undefined) {
        console.log(response.data['Error'])
        if (failedCallback !== null && failedCallback !== undefined) failedCallback(webErrorHandle(response.data['Error']))
    } else if  (response.data['Redirect'] !== '' && response.data['Redirect'] !== null && response.data['Redirect'] !== undefined) {
        window.location.href = response.data['Redirect'];
    } else if  (response.data['Data'] !== '' && response.data['Data'] !== null && response.data['Data'] !== undefined) {
        if (successedCallback !== null && successedCallback !== undefined) successedCallback(response)
    }
}

function webErrorHandle(data) {
    console.log(data)
    let sdata = '' + data;
    let code = '';
    if (sdata.indexOf('-') > -1) {
        code = sdata.split('-')[1];
        sdata = sdata.split('-')[0];
    }
    console.log(sdata)

    // httpGet('https://fanyi.youdao.com/translate?&doctype=json&type=AUTO&i=' + sdata)
    // .then(function (response) {
    //     console.log(response.data)
    //     if (response.data['errorCode'] !== undefined && response.data['errorCode'] !== null && response.data['errorCode'] !== 0) {
    //         if (response.data['translateResult'].length > 0) {
    //             if (response.data['translateResult'][0].length > 0) {
    //                 if (response.data['translateResult'][0][0]['src'] !== undefined && response.data['translateResult'][0][0]['src'] !== null && response.data['translateResult'][0][0]['src'] === sdata) {
    //                     sdata = response.data['translateResult'][0][0]['tgt'];
    //                 }
    //             }
    //         }
    //     }
    //     if (callback !== null) callback.info(sdata);
    // })
    // .catch(function (response) {
    //     if (callback !== null) callback.info(sdata);
    // });
    // if (callback !== null && callback !== undefined) callback.$Message.error({
    //     content: code === '' ? sdata:sdata + ', code: ' + code,
    //     duration: 3,
    //     closable: true
    // });
    // return sdata;
    return {
        content: code === '' ? sdata:sdata + ', code: ' + code,
        duration: 3,
        closable: true
    }
}

export {webResponse}