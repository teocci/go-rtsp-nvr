const wsProtocol = location.protocol === 'https:' ? 'wss' : 'ws'
const playerElement = document.getElementById('player')
const url = `${wsProtocol}://${location.host}/stream/${uuid}/channel/${channel}/mse?uuid=${uuid}&channel=${channel}`

let mseQueue = [], mseStreamingStarted = false
let mseSourceBuffer

function startPlay() {
    console.log('On startPlay')
    console.log({url})
    const mse = new MediaSource()

    playerElement.src = window.URL.createObjectURL(mse)
    mse.onsourceopen = () => {
        const ws = new WebSocket(url)
        ws.binaryType = "arraybuffer"

        ws.onopen = event => {
            console.log('Connect to ws')
        }

        ws.onmessage = function(event) {
            let data = new Uint8Array(event.data)
            if (data[0] === 9) {
                const decodedData = data.slice(1)
                const mimeCodec = window.TextDecoder ? new TextDecoder("utf-8").decode(decodedData) : Utf8ArrayToStr(decodedData)

                mseSourceBuffer = mse.addSourceBuffer(`video/mp4; codecs="${mimeCodec}"`)
                mseSourceBuffer.mode = 'segments'
                mseSourceBuffer.addEventListener('updateend', pushPacket)
            } else {
                readPacket(event.data)
            }
        }
    }
}

function pushPacket() {
    if (!mseSourceBuffer.updating) {
        if (mseQueue.length > 0) {
            const packet = mseQueue.shift()
            mseSourceBuffer.appendBuffer(packet)
        } else {
            mseStreamingStarted = false
        }
    }

    if (playerElement.buffered.length > 0) {
        if (typeof document.hidden !== "undefined" && document.hidden) {
            //no sound, browser paused video without sound in background
            playerElement.currentTime = playerElement.buffered.end((playerElement.buffered.length - 1)) - 0.5
        }
    }
}

function readPacket(packet) {
    if (!mseStreamingStarted) {
        mseSourceBuffer.appendBuffer(packet)
        mseStreamingStarted = true
        return
    }

    mseQueue.push(packet)

    if (!mseSourceBuffer.updating) {
        pushPacket()
    }
}

playerElement.onloadeddata = () => {
    playerElement.play()
    const browser = browserDetector()
    if (!browser.safari) {
        makePic()
    }
}

//fix stalled video in safari
playerElement.addEventListener('pause', () => {
    if (playerElement.currentTime > playerElement.buffered.end((playerElement.buffered.length - 1))) {
        playerElement.currentTime = playerElement.buffered.end((playerElement.buffered.length - 1)) - 0.1
        playerElement.play()
    }
})

playerElement.onerror = () => {
    console.log('video_error')
}

window.onload = () => {
    startPlay()
}