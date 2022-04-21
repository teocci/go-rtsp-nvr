if (typeof jQuery === 'undefined') {
  throw new Error('need jQuery');
}

(function($) {
  var IpeyePlayer = function(element, options) {
    this.$element = null;
    this.options = null;
    this.ms = null;
    this.ws = null;
    this.wsmesagging = false;
    this.video = null;
    // this.url_devcode_info = "index.php?route=page_play_ajax&new_websocket";
    // this.url_flash = 'index.php?route=page_play_ajax';
    this.url_devcode_info = "//ipeye.ru/webs/stream_info.php";
    this.url_flash = '//ipeye.ru/webs/stream_info.php?flash';
    this.sound = false;
    this.mimeType = 'video/mp4';
    this.sourceBuffer = null;
    this.queue = [];
    this.streamingStarted = false;
    this.streamUrl = null;
    this.browser = null;
    this.canvas_screenshot = null;
    this.averageBuffer = [];
    this.ctx_screenshot = null;
    this.paused = false;
    this.stream_name = '';
    this.hls = false;
    this.ptz_control = false;
    this.params = '';
    this.setTimeoutId = null;
    this.canvas = null;
    this.ctx = null;
    this.window_bind = null;
    this.bitrate = [];
    this.data_size = 0;
    this.gop_decode = null;
    this.segments_count = 0;
    this.messageTimeout = null;
    this.zoomenabled = false;
    this.zoom_bind = null;
    this.scale = 1;
    this.mousedown = false;
    this.down_pos_x = 0;
    this.down_pos_y = 0;
    this.translate_x = 0;
    this.translate_y = 0;
    this.webrtc=null;
    this.webrtcSendChannel=null;
    this.webrtcSendChannelInterval=null;
    this.error_counter=0;
    this.webrtcMediaStream=null;
    this.init(element, options);
    return this;
  }

  IpeyePlayer.DEFAULTS = {
    devcode: '',
    logo: {
      img: '//ipeye.ru/img/logo/onsite2.png',
      url: '//ipeye.ru'
    },
    autoplay: true,
    muted: true,
    debug: false,
    play_info: false,
    controls: true,
    ptz_control: false,
    onloadvideo: null,
    onbufferedfideo: null,
    onrendercontrol: null,
    onplay: null,
    wsurl: null,
    mobile: false,
    flash_only: false,
    oncanvasresize: false,
    disable_click_to_play: false,
    zoomable: false,
    onFullscreenEvent: null,
    webtrcplay: false,
    poster:true,
    poster_url:null
  }

  IpeyePlayer.prototype.init = function(element, options) {
    this.options = this.getOptions(options);
    this.browser = this.browserDetect();
    //console.log(this.browser);
    this.logging('[init]=>Начата инициализация плеера');
    this.$element = $(element);
    this.$element.empty();
    this.$element.addClass('ipeye-player');

    this.$element.append('<div class="spinner-three-bounce hidden"><div style="animation-delay: -1.5s;"></div><div data-bounce2=""></div><div data-bounce3=""></div></div>');


    if (this.options.autoplay) {
      this.params += ' autoplay ';
    } else {
      this.paused = true;
    }
    if (this.options.muted) {
      this.params += ' muted ';
    }
    if(this.options.poster){
      if(this.options.poster_url!=null){
        this.params += ' poster="'+this.options.poster_url+'" ';
      }else{
        this.params += ' poster="//api.ipeye.ru/v1/stream/poster/1/'+this.options.devcode+'/img.jpeg" ';
      }

    }
    this.$element.append('<video  playsinline="true"  ' + this.params + ' type="video/mp4"></video>');
    this.video = this.$element.find('video')[0];
    this.$element.append('<div class="message-wrapper hidden"><span class="before"></span><span class="inner"></span></div>');
    /*
     * ie slice fix
     */
    if (!Uint8Array.prototype.slice) {
      Object.defineProperty(Uint8Array.prototype, 'slice', {
        value: function(begin, end) {
          return new Uint8Array(Array.prototype.slice.call(this, begin, end));
        }
      });
    }
    /****************************************/

    if (!window.MediaSource || this.options.flash_only || this.options.webtrcplay) {
      if(this.webrtcsupport()){
        this.webrtcMediaStream=new MediaStream();
        this.addVideoListeners();
        this.$element.append('<canvas class="screenshot_canvas" style="display:none"></canvas>');
        this.canvas_screenshot = this.$element.find('canvas.screenshot_canvas')[0];
        this.ctx_screenshot = this.canvas_screenshot.getContext('2d');

        if (this.options.logo) {
          var href = '#';
          if (!!this.options.logo.url) {
            href = this.options.logo.url;
          }
          this.$element.append('<div class="logo"><a href="' + href + '" target="_blank"><img src="' + this.options.logo.img + '" /></a></div>')
        }
        if (this.options.play_info) {
          this.$element.append('<div class="play-info"></div>');
        }

        this.$element.append('<canvas class="draw_canvas" style="display:none"></canvas>');
        this.canvas = this.$element.find('canvas.draw_canvas')[0];
        this.ctx = this.canvas.getContext('2d');

        //this.element.addEventListener('mousedown', this.onMouseDown.bind(this))
        var _this = this;
        if (!this.options.disable_click_to_play) {
          this.$element.on('click', '.draw_canvas', function(e) {

            if (!!_this.video) {
              if (_this.paused) {
                _this.paused = false;
                _this.$element.find('.media-control-icons[data-action="playpause"]').html(_this.svgIcons('pause'));
              } else {
                _this.paused = true;
                _this.$element.find('.media-control-icons[data-action="playpause"]').html(_this.svgIcons('play'));
              }
            }

          });
        }

        this.window_bind = this.canvasReposition.bind(this);
        window.addEventListener('resize', this.window_bind);
        if (_this.options.controls) {

          if (!_this.$element.find('.media-control').length) {
            _this.addButtons();
          }

        }
        this.webrtcStream();

      }else  if (this.video.canPlayType('application/vnd.apple.mpegurl')) {
        this.hls = true;
        this.hlsStream();
      } else {
        this.playerShowMessage('Ваш браузер не поддерживает потоковое воспроизведение<br> используйте другой браузер');
      }

    } else {

      this.addVideoListeners();
      this.$element.append('<canvas class="screenshot_canvas" style="display:none"></canvas>');
      this.canvas_screenshot = this.$element.find('canvas.screenshot_canvas')[0];
      this.ctx_screenshot = this.canvas_screenshot.getContext('2d');

      if (this.options.logo) {
        var href = '#';
        if (!!this.options.logo.url) {
          href = this.options.logo.url;
        }
        this.$element.append('<div class="logo"><a href="' + href + '" target="_blank"><img src="' + this.options.logo.img + '" /></a></div>')
      }
      if (this.options.play_info) {
        this.$element.append('<div class="play-info"></div>');
      }

      this.$element.append('<canvas class="draw_canvas" style="display:none"></canvas>');
      this.canvas = this.$element.find('canvas.draw_canvas')[0];
      this.ctx = this.canvas.getContext('2d');

      //this.element.addEventListener('mousedown', this.onMouseDown.bind(this))
      var _this = this;
      if (!this.options.disable_click_to_play) {
        this.$element.on('click', '.draw_canvas', function(e) {

          if (!!_this.video) {
            if (_this.paused) {
              _this.paused = false;
              _this.$element.find('.media-control-icons[data-action="playpause"]').html(_this.svgIcons('pause'));
            } else {
              _this.paused = true;
              _this.$element.find('.media-control-icons[data-action="playpause"]').html(_this.svgIcons('play'));
            }
          }

        });
      }

      this.window_bind = this.canvasReposition.bind(this);
      window.addEventListener('resize', this.window_bind);

      this.getInfoAndStart();



    }

  }

  IpeyePlayer.prototype.getDefaults = function() {
    return IpeyePlayer.DEFAULTS;
  }

  IpeyePlayer.prototype.getOptions = function(options) {
    options = $.extend({}, this.getDefaults(), options);
    return options;
  }

  IpeyePlayer.prototype.messToCon = function(message) {
    if (this.options.play_info) {
      this.$element.find('.play-info').html(message);
    }
  }
  IpeyePlayer.prototype.hlsStream = function() {
    var _this = this;
    $.ajax({
      url: _this.url_devcode_info,
      type: 'GET',
      beforeSend: function() {
        _this.playerShowMessage('Получаю даннные о потоке ...');
      },
      data: 'devid=' + _this.options.devcode,
      success: function(data) {
        //console.log(data);
        _this.playerHideMessage();
        if (!!data) {
          try {
            data = JSON.parse(data);
            if (data.status == 1) {
              // if (data.ondemand) {
              //   _this.playerShowMessage('Вашe устройство не может воспроизводить видео на бесплатном тарифе. <br> Смените устройство или тариф');
              //   return;
              // } else {
              _this.logging('[ajax ] => Данные получены,  поток онлайн, можно запускать');
              //http://171.25.232.15/api/v1/stream/f7404d80f4a14ffe93ba72c72523920d/hls/index.m3u8
              _this.$element.find('video').remove();
              _this.$element.append('<video src="//' + data.server + '/api/v1/stream/' + _this.options.devcode + '/hls/index.m3u8"  muted autoplay="true" controls playsinline></video>');
              _this.video = _this.$element.find('video')[0];
              _this.addVideoListeners();
              if (_this.options.logo) {
                var href = '#';
                if (!!_this.options.logo.url) {
                  href = _this.options.logo.url;
                }
                _this.$element.append('<div class="logo"><a href="' + href + '" target="_blank" ><img src="' + _this.options.logo.img + '" /></a></div>')
              }
              _this.$element.append('<canvas class="draw_canvas" style="display:none"></canvas>');
              _this.canvas = _this.$element.find('canvas.draw_canvas')[0];
              _this.ctx = _this.canvas.getContext('2d');

              _this.window_bind = _this.canvasReposition.bind(_this);
              window.addEventListener('resize', _this.window_bind);
              // }



            } else {
              if (!!data.message) {
                _this.playerShowMessage(data.message);
              } else {
                _this.playerShowMessage('Поток выключен');
              }

            }
            if (!!data.stream_name) {
              _this.stream_name = data.stream_name;
            }
          } catch (e) {
            console.log(e);
            _this.playerShowMessage('Ошибка :(');
          }
        } else {
          _this.playerShowMessage('Нет такого потока');
        }
      },
      error: function() {
        _this.playerShowMessage('Ошибка ajax :(');
      }
    })
  }

  IpeyePlayer.prototype.webrtcStream = function() {
    var _this = this;
    $.ajax({
      url: _this.url_devcode_info,
      type: 'GET',
      beforeSend: function() {
        _this.playerShowMessage('Получаю даннные о потоке WEBRTC ...');
      },
      data: 'devid=' + _this.options.devcode,
      success: function(data) {
        //console.log(data);
        _this.playerHideMessage();
        if (!!data) {
          try {
            data = JSON.parse(data);
            if (data.status == 1) {
              _this.logging('[ajax ] => Данные получены,  поток онлайн, можно запускать');
              _this.startWebRtcPlay(data.server,_this.options.devcode);

            } else {
              if (!!data.message) {
                _this.playerShowMessage(data.message);
              } else {
                _this.playerShowMessage('Поток выключен');
              }

            }
            if (!!data.stream_name) {
              _this.stream_name = data.stream_name;
            }
          } catch (e) {
            console.log(e);
            _this.playerShowMessage('Ошибка :(');
          }
        } else {
          _this.playerShowMessage('Нет такого потока');
        }
      },
      error: function() {
        _this.playerShowMessage('Ошибка ajax :(');
      }
    })
  }



  IpeyePlayer.prototype.getflashplayer = function() {
    var _this = this;
    $.ajax({
      url: _this.url_flash,
      data: 'devid=' + _this.options.devcode,
      success: function(data) {

        if (data) {
          _this.$element.html(data);
          //console.log(data);
        }
      }
    })
  }
  IpeyePlayer.prototype.createMedia = function() {

    var _this = this;
    _this.sourceBuffer = null;
    _this.ms = new MediaSource();
    //  _this.logging('beforeonsourceopen');
    _this.ms.addEventListener('sourceopen', function() {
      //_this.logging('onsourceopen');
      _this.websocketConnect();
    });

    _this.ms.addEventListener('sourceended', function() {
      _this.logging('sourceended');

    }, false);

    _this.ms.addEventListener('sourceclose', function() {
      _this.logging('sourceclose');

    }, false);

    _this.video.src = window.URL.createObjectURL(_this.ms);
  }

  IpeyePlayer.prototype.getInfoAndStart = function() {
    var _this = this;
    this.logging('[ajax]=>Начата запрос данных о потоке');
    if (this.options.wsurl != null) {
      //console.log(this.options.wsurl);
      _this.logging('[ajax] => ajax не понадобился - передан прямой урл');
      _this.streamUrl = _this.options.wsurl;
      _this.createMedia();

    } else {
      $.ajax({
        url: _this.url_devcode_info,
        type: 'GET',
        beforeSend: function() {
          _this.playerShowMessage('Получаю даннные о потоке ...');
        },
        data: 'devid=' + _this.options.devcode,
        success: function(data) {
          _this.playerHideMessage();

          if (!!data) {
            try {
              data = JSON.parse(data);
              if (data.status == 1) {

                _this.logging('[ajax => ] есть сервер и поток онлайн, можно запускать');
                if (!!data.mobile && data.mobile == 1 && _this.options.mobile) {
                  _this.logging('[mobile => ] на наличие доп потока ' + _this.options.devcode + ' сервер ответил: ' + data.mobile);
                  _this.options.mobile = true;

                } else {
                  _this.options.mobile = false;
                }

                _this.streamUrl = _this.getStreamUrl(data.server);

                if (data.ptz == 1 && _this.options.ptz_control) {
                  _this.ptz_control = true;
                }
                //_this.logging('_this.createMedia();');
                _this.createMedia();

              } else {
                if (!!data.message) {
                  _this.playerShowMessage(data.message);
                } else {
                  _this.playerShowMessage('Поток выключен');
                }

              }
              if (!!data.stream_name) {
                _this.stream_name = data.stream_name;
              }
              if (!!data.stopLive) {
                _this.messageTimeout = setTimeout(function() {
                  _this.blurplayer(data.stopLive.message);
                }, data.stopLive.timeout);
              }
            } catch (e) {
              _this.logging(e);
              //console.log(e);
              _this.playerShowMessage('Ошибка :(');
            }
          } else {
            _this.playerShowMessage('Нет такого потока');
          }
        },
        error: function() {
          _this.playerShowMessage('Ошибка ajax :(');
        }
      })
    }

  }

  IpeyePlayer.prototype.protocol = function() {
    protocol = 'ws';
    if (location.protocol.indexOf('s') >= 0) {
      protocol = 'wss';
    }
    return protocol;
  }

  IpeyePlayer.prototype.getStreamUrl = function(server) {
    if (this.options.mobile) {
      return this.protocol() + '://' + server + '/ws/mp4/live?name=' + this.options.devcode + '_mobile';
    } else {
      return this.protocol() + '://' + server + '/ws/mp4/live?name=' + this.options.devcode;
    }

    //  console.log(this.protocol() + '://171.25.233.50:8080/ws/mp4/live?name=test');
    //return this.protocol() + '://171.25.233.50:8080/ws/mp4/live?name=test';


  }

  IpeyePlayer.prototype.playerShowMessage = function(message) {
    this.$element.find('.message-wrapper .inner').html(message);
    this.$element.find('.message-wrapper').removeClass('hidden');
  }

  IpeyePlayer.prototype.playerHideMessage = function() {
    this.$element.find('.message-wrapper').addClass('hidden');
  }

  IpeyePlayer.prototype.playerShowControl = function() {
    this.$element.find('.media-control').removeClass('media-control-hide');
  }
  IpeyePlayer.prototype.playerHideControl = function() {
    this.$element.find('.media-control').addClass('media-control-hide');
  }

  IpeyePlayer.prototype.websocketConnect = function(url) {
    this.ws = new WebSocket(this.streamUrl);
    this.ws.binaryType = "arraybuffer";
    this.websocketAddListeners();
  }

  IpeyePlayer.prototype.websocketAddListeners = function() {
    var _this = this;
    _this.ws.onopen = function(event) {
      _this.logging('[ws]=>Соединение установлено');
      _this.playerHideMessage();
    }

    _this.ws.onclose = function(event) {
      if (event.wasClean) {
        _this.logging('[ws]=>Соединение закрыто чисто');
      } else {
        _this.logging('[ws]=>Обрыв соединения');
        _this.$element.find('.spinner-three-bounce').addClass('hidden');
        _this.playerShowMessage('Потеря соединения. Попытка переподключения через <b class="reconnect_timer" >5</b> секунд');

        _this.websocketRemoveListeners();

        if (!_this.wsmesagging && _this.options.mobile) {
          //запрошен мобильный поток но сервер не ответил совсем ничего
          _this.playerShowMessage('Был запрошен дополнительный поток, но воспроизвести его не удалось, проверьте правильно ли вы указали адрес дополнительного потока в настройках камеры<br> Будет воспроизведен основной поток через <b class="reconnect_timer" >5</b> секунд');
          _this.options.mobile = false;
          _this.streamUrl = _this.streamUrl.slice(0, _this.streamUrl.indexOf('_mobile'));
        }
        _this.reconnectTimer();
        _this.setTimeoutId = setTimeout(function() {
          _this.websocketConnect();
        }, 5000)
      }
      _this.logging('Код: ' + event.code + ' причина: ' + event.reason);
    };

    _this.ws.onerror = function(error) {
      _this.logging("Ошибка " + error.message);
    };

    _this.ws.onmessage = function(event) {
      _this.wsmesagging = true; //пришло хотя бы одно сообщение от сервера
      if(_this.error_counter<3){
        _this.packetCatcher(event.data);
      }

    }

  }
  IpeyePlayer.prototype.reconnectTimer = function() {
    var _this = this;
    var count = parseInt(this.$element.find('.reconnect_timer').html());
    //console.log(count);
    if (count > 0) {
      setTimeout(function() {
        _this.$element.find('.reconnect_timer').html((count - 1));
        _this.reconnectTimer();
      }, 1000)
    } else {
      return false;
    }

  }

  IpeyePlayer.prototype.websocketRemoveListeners = function() {
    var _this = this;
    if (!!_this.ws) {
      _this.ws.onerror = null;
      _this.ws.onopen = null;
      _this.ws.onmessage = null;
      _this.ws.onclose = null;

    }
  }
  IpeyePlayer.prototype.packetCatcher = function(packet) {
    var _this = this;
    var data = new Uint8Array(packet);
    //первый пакет приносит

    _this.data_size += (data.byteLength);
    //console.log((_this.data_size / 1024 / 1024).toFixed(2) + 'Mb');
    if (data[0] == 6) {
      var mimeCodec = '';
      var decoded_arr;
      try {
        decoded_arr = data.slice(1);
      } catch (e) {
        _this.logging(e);
        //console.log(e);
      }
      if (window.TextDecoder) {
        mimeCodec = new TextDecoder("utf-8").decode(decoded_arr);
      } else {
        mimeCodec = _this.Utf8ArrayToStr(decoded_arr);
      }
      var temp_mime = mimeCodec;
      if (temp_mime.indexOf(',') >= 0) {
        temp_mime = mimeCodec.split(',');
        if (temp_mime.length > 2) {
          //mimeCodec = temp_mime[0];
        }
      }
      this.logging('[packetCatcher]=>Кодек ' + mimeCodec);
      if ((mimeCodec.indexOf(',')) >= 0) {
        this.sound = true;
      }
      mimeCodec = 'video/mp4; codecs="' + mimeCodec + '"';
      //console.log('MediaSource' in window, MediaSource.isTypeSupported(mimeCodec))
      if ('MediaSource' in window && MediaSource.isTypeSupported(mimeCodec)) {


        if (this.sourceBuffer == null) {

          this.sourceBuffer = this.ms.addSourceBuffer(mimeCodec);
          if (this.browser.chrome && (this.browser.linux || this.browser.android)) {
            this.sourceBuffer.mode = "segments";
          } else {
            //this.sourceBuffer.mode = "segments";
            this.sourceBuffer.mode = "sequence";
            //console.log('sequence');
          }

          this.sourceBufferlistener();

          if (this.options.controls) {

            if (!this.$element.find('.media-control').length) {
              this.addButtons();
            }

          }
        }

      } else {
        _this.ws.close(1000);
        _this.$element.find('.spinner-three-bounce').addClass('hidden');
        _this.playerShowMessage('Ваш браузер не поддерживает кодек вашего устройства, смените кодек на H.264 или на H.265X');

      }
      _this.segments_count = 0;
    } else {

      //console.log(packet);
      if (!this.paused) {

        this.pushPacket(packet);
        if (typeof document.hidden !== "undefined" && document.hidden) {
          //на спрятанной вкладке плей не дергаем
        } else {
          if (this.video.paused) {
            this.video.play();
          }
        }

      } else {
        if (!this.video.paused) {
          this.video.pause();
        }
        if (this.video.readyState != 4) {
          this.pushPacket(packet);
        }
      }

      // if (this.video.readyState == 4 && this.video.paused) {
      //
      // }
      this.segments_count++;
    }
  }

  IpeyePlayer.prototype.jumpToEnd = function() {
    var _this = this;
    if (_this.sourceBuffer.buffered.length > 0) {
      var range = _this.sourceBuffer.buffered.length - 1;
      if ((_this.sourceBuffer.buffered.end(range) - _this.video.currentTime) >= _this.getAverageBuffer()) {

        _this.video.currentTime = _this.sourceBuffer.buffered.end(range) - (_this.getAverageBuffer()*1.5);
      }

    }
  }

  IpeyePlayer.prototype.sourceBufferlistener = function() {
    var _this = this;
    this.sourceBuffer.addEventListener("updateend", function() {
      if (_this.sourceBuffer.buffered.length > 0) {
        var range = _this.sourceBuffer.buffered.length - 1;
        if (typeof _this.options.onbufferedfideo == 'function') {
          _this.options.onbufferedfideo((_this.sourceBuffer.buffered.end(range) - _this.video.currentTime).toFixed(3));
        }
        var log = '<p>Загружено сегментов: <span class="pull-right">' + (range + 1) + '</span></p>';
        log += '<p>Длительность видео: <span class="pull-right">' + _this.sourceBuffer.buffered.end(range).toFixed(3) + ' с.</span></p>';
        log += '<p>Буфер для декодирования GOP: <span class="pull-right">' + (_this.sourceBuffer.buffered.end(range) - _this.video.currentTime).toFixed(3) + ' с.</span></p>';

        log += '<p>Загружено данных: <span class="pull-right">' + (_this.data_size / 1024 / 1024).toFixed(2) + ' Mb' + '</span></p>';
        log += '<p>canplaythrough time: <span class="pull-right">' + _this.getAverageBuffer().toFixed(2) + ' c.' + '</span></p>';
        var videoPlaybackQuality = _this.video.getVideoPlaybackQuality;

        if (videoPlaybackQuality && typeof videoPlaybackQuality === typeof Function) {
          log += '<p>Dropped frames: <span class="pull-right">' + _this.video.getVideoPlaybackQuality().droppedVideoFrames + '</span></p>';
          log += '<p>Corrupted frames: <span class="pull-right">' + _this.video.getVideoPlaybackQuality().corruptedVideoFrames + '</span></p>';
          log += '<p>Total frames: <span class="pull-right">' + _this.video.getVideoPlaybackQuality().totalVideoFrames + '</span></p>';

        } else if (_this.video.webkitDroppedFrameCount) {
          log += '<p>Dropped frames: <span class="pull-right">' + _this.video.webkitDroppedFrameCount + '</span></p>';
        }


        _this.messToCon(log);


        if (typeof document.hidden !== "undefined" && document.hidden) {
          //спрятали вкладку

          if (!_this.sound) {
            //звука нет будет отставать будем подгонять
            _this.video.currentTime = _this.sourceBuffer.buffered.end(range) - _this.getAverageBuffer();


          }
        }
        //для всех - спрятанных или нет если текущее отстает от среднего больше чем в 2 раза то прыгнем в конец
        if ((_this.sourceBuffer.buffered.end(range) - _this.video.currentTime) >= (Math.abs(_this.getAverageBuffer()) * 3)) {

          _this.jumpToEnd();
        }

        //перепрыгнем через пустые места
        _this.checkRanges();

      }

      _this.loadPacket();
    });


    this.sourceBuffer.addEventListener("error", function(e) {
      _this.logging(e);
      this.ws.close(1000);
      this.playerShowMessage('Непредвиденная ошибка потока, передача остановлена для возобновления перезагрузите страницу');
    });
  }
  IpeyePlayer.prototype.checkRanges = function() {
    if (this.sourceBuffer.buffered.length > 0) {
      var range = this.sourceBuffer.buffered.length - 1;
      if (this.video.currentTime < this.sourceBuffer.buffered.start(range)) {
        this.video.currentTime = this.sourceBuffer.buffered.end(range) - this.getAverageBuffer();
      }
    }
  }
  IpeyePlayer.prototype.pushPacket = function(packet) {
    var view = new Uint8Array(packet);

    if (!this.streamingStarted && !this.sourceBuffer.updating) {
      try {
        this.sourceBuffer.appendBuffer(view);
        this.streamingStarted = true;
      } catch (e) {
        this.logging(e);
        this.ws.close(1000);
        this.playerShowMessage('Непредвиденная ошибка потока, передача остановлена для возобновления перезагрузите страницу');
      }

      return;
    }

    this.queue.push([].slice.call(view));

    if (!this.sourceBuffer.updating) {
      this.loadPacket();
    }
  }
  IpeyePlayer.prototype.loadPacket = function() {


    if (!this.sourceBuffer.updating) {

      if (this.queue.length > 0) {
        var view = new Uint8Array(this.queue.shift());
        try {
          this.sourceBuffer.appendBuffer(view);
        } catch (e) {
          this.logging(e);

          this.ws.close(1000);
          this.playerShowMessage('Непредвиденная ошибка потока, передача остановлена для возобновления перезагрузите страницу');
        }

      } else {

        this.streamingStarted = false;
      }
    } else {

    }
  }

  IpeyePlayer.prototype.getAverageBuffer = function() {
    if (this.gop_decode != null) {
      return this.gop_decode;
    } else {
      return 2;
    }
  }

  IpeyePlayer.prototype.addVideoListeners = function() {
    var _this = this;

    _this.video.oncanplaythrough = function() {
      _this.$element.find('.spinner-three-bounce').addClass('hidden');

      if (!_this.hls&&_this.webrtc==null) {
        if (_this.sourceBuffer.buffered.length > 0) {
          if (_this.gop_decode == null) {
            _this.gop_decode = (_this.sourceBuffer.buffered.end(0) - _this.video.currentTime);
          }

        }
      }

    }

    _this.video.onplay = function() {
      _this.$element.find('.spinner-three-bounce').addClass('hidden');
      if (typeof _this.options.onplay == 'function') {

        _this.options.onplay();
      }
    }

    _this.video.oncanplay = function() {

      _this.$element.find('.spinner-three-bounce').addClass('hidden');
      if (_this.gop_decode == null && _this.webrtc==null) {
        _this.gop_decode = (_this.sourceBuffer.buffered.end(0) - _this.video.currentTime);

      }
    }
    _this.video.onerror = function(e) {
      _this.error_counter++;

      if(_this.error_counter>2){
        _this.playerShowMessage('Критическая масса ошибок!Попробуйте перезагрузить страницу либо обратиться в техподдержку');
        _this.logging('[video error] ==> code: ' + e.target.error.code + ' message: ' + e.target.error.message);
        return;
      }

      if (_this.hls) {
        _this.destroy();
        _this.hlsStream();
      }else if (_this.webrtc!=null) {
        _this.logging('[video error] ==> webrtcerrcode: ' + e.target.error.code + ' message: ' + e.target.error.message);
        _this.webrtcSendChannel.onclose=null;
        _this.webrtcSendChannel.onopen=null;
        clearInterval(_this.webrtcSendChannelInterval);
        _this.webrtc.close();
        _this.video.removeAttribute('src');
        _this.video.load();
        _this.$element.find('video').remove();
        _this.$element.append('<video playsinline="true"  muted autoplay="true" type="video/mp4"></video>');
        _this.video = _this.$element.find('video')[0];
        _this.addVideoListeners();
        _this.webrtcStream();
      } else {
        _this.logging('[video error] ==> code: ' + e.target.error.code + ' message: ' + e.target.error.message);

        //console.log('error', e, _this.video.error);
        _this.video.pause();
        _this.ws.close(1000);
        _this.websocketRemoveListeners();
        _this.video.removeAttribute('src');
        _this.video.load();
        _this.$element.find('video').remove();
        _this.$element.append('<video playsinline="true"  muted autoplay="true" type="video/mp4"></video>');
        _this.video = _this.$element.find('video')[0];
        _this.addVideoListeners();
        _this.createMedia();
      }



    };

    _this.video.onloadeddata = function() {
      _this.$element.find('canvas.draw_canvas').css("display", "block");

      _this.canvasReposition();
      if (typeof _this.options.onloadvideo == 'function') {
        _this.options.onloadvideo(_this.video.videoWidth, _this.video.videoHeight, _this.stream_name)
      }
    }

    _this.video.addEventListener('loadstart', function() {
      _this.$element.find('.spinner-three-bounce').removeClass('hidden');
    });

    _this.video.addEventListener('stalled', function() {
      console.log('stalled');
    });

    _this.video.addEventListener('waiting', function() {

      _this.$element.find('.spinner-three-bounce').removeClass('hidden');
    });
  }

  IpeyePlayer.prototype.addButtons = function() {
    var _this = this;
    this.$element.append('<div class="media-control media-control-hide">' +
      '<div class="media-control-background"></div>' +
      '<div class="media-control-layer">' +
      '<div class="media-control-left-panel"></div>' +
      '<div class="media-control-center-panel"><div class="media-control-icons" data-action="restart">LIVE</div></div>' +
      '<div class="media-control-right-panel"></div>' +
      '</div></div>');
    //Звук
    if (this.sound) {
      this.$element.find('.media-control-right-panel').append('<div class="media-control-icons" data-action="sound">' + this.svgIcons('muted') + '</div>');
    }

    if (this.options.zoomable) {
      this.$element.find('.media-control-right-panel').append('<div class="media-control-icons" data-action="zoom">' + this.svgIcons('zoom') + '</div>');
    }
    //полный экран
    if (this.fullscreenEnabled()) {
      this.$element.find('.media-control-right-panel').append('<div class="media-control-icons" data-action="expand">' + this.svgIcons('expand') + '</div>');
    }
    if (this.paused) {
      this.$element.find('.media-control-left-panel').append('<div class="media-control-icons" data-action="playpause">' + this.svgIcons('play') + '</div>');
    } else {
      this.$element.find('.media-control-left-panel').append('<div class="media-control-icons" data-action="playpause">' + this.svgIcons('pause') + '</div>');
    }

    //скриншот
    if (!this.browser.safari) {
      this.$element.find('.media-control-left-panel').append('<div class="media-control-icons" data-action="screenshot">' + this.svgIcons('screenshot') + '</div>');
    }

    if (_this.ptz_control) {
      //console.log('pts')
      _this.$element.find('.control_panel').remove();
      _this.$element.find(".media-control-center-panel").html('<div class="control_panel" code="' + _this.options.devcode + '"><button class="control_panel_btn" data-action="left"><i class="fa fa-caret-left"></i></button><button class="control_panel_btn" data-action="up"><i class="fa fa-caret-up"></i></button><button class="control_panel_btn" data-action="down"><i class="fa fa-caret-down"></i></button><button class="control_panel_btn" data-action="right"><i class="fa fa-caret-right"></i></button><button class="control_panel_btn"  data-action="zoomin"><i class="fa fa-search-plus " ></i></button><button class="control_panel_btn" data-action="zoomout"><i class="fa fa-search-minus "></i></button></div>');
    }


    this.$element.on('click', '.media-control-icons', function() {

      switch ($(this).data('action')) {
        case 'expand':
          if (document.fullscreenElement || document.webkitCurrentFullScreenElement ||
            document.webkitFullscreenElement || document.mozFullScreenElement || document.msFullscreenElement) {
            if (typeof _this.options.onFullscreenEvent == 'function') {
              _this.options.onFullscreenEvent(0);
            }
            _this.closeFullscreen();

          } else {
            if (typeof _this.options.onFullscreenEvent == 'function') {
              _this.options.onFullscreenEvent(1);
            }
            _this.openFullscreen(_this.$element[0]);

          }



          break;
        case 'zoom':
          if (_this.zoomenabled) {
            //console.log('off zoom')
            _this.zoomenabled = false;
            _this.$element.find('[data-action="zoom"]').css('filter','');
            _this.disableZoom();
          } else {
            //console.log('on zoom')
            _this.zoomenabled = true;
            _this.$element.find('[data-action="zoom"]').css('filter','invert(1)');;
            _this.enableZoom();

          }

          break;
        case 'sound':
          if (_this.video.muted) {
            _this.video.muted = false;
            $(this).html(_this.svgIcons('speaker'));

          } else {
            _this.video.muted = true;
            $(this).html(_this.svgIcons('muted'));
          }


          break;

        case 'restart':
          _this.restartStream();



          break;
        case 'playpause':

          if (_this.paused) {
            _this.paused = false;
            _this.video.play();
            $(this).html(_this.svgIcons('pause'));
          } else {
            _this.paused = true;
            _this.video.pause();
            $(this).html(_this.svgIcons('play'));
          }
          break;
        case 'screenshot':

          _this.screenshot();

          break;
        default:

      }
      return;
    });
    this.$element.on('mouseover', function() {
      _this.playerShowControl();

    })
    this.$element.on('mouseout', function() {
      _this.playerHideControl();

    })
    // this.$element.on('click', function(e) {
    //   console.log(e.target);
    //
    // });

    if (typeof _this.options.onrendercontrol == 'function') {
      //console.log('render');
      _this.options.onrendercontrol();
    }
  }
  IpeyePlayer.prototype.enableZoom = function() {
    // включили зумм - навесить обработчики на елемент
    this.zoom_bind = this.zoomHandler.bind(this);
    if ('onwheel' in document) {
      this.$element[0].addEventListener("wheel", this.zoom_bind);
    } else if ('onmousewheel' in document) {
      this.$element[0].addEventListener("mousewheel", this.zoom_bind);
    } else {
      this.$element[0].addEventListener("MozMousePixelScroll", this.zoom_bind);
    }
    this.$element[0].addEventListener("mousemove", this.zoom_bind);
    this.$element[0].addEventListener("mousedown", this.zoom_bind);
    this.$element[0].addEventListener("mouseup", this.zoom_bind);

  }

  IpeyePlayer.prototype.disableZoom = function() {
    //отключили зумм - вернуть в нормильный размер и отключить обработчики
    if (this.zoom_bind != null) {
      if ('onwheel' in document) {
        this.$element[0].removeEventListener("wheel", this.zoom_bind);
      } else if ('onmousewheel' in document) {
        this.$element[0].removeEventListener("mousewheel", this.zoom_bind);
      } else {
        this.$element[0].removeEventListener("MozMousePixelScroll", this.zoom_bind);
      }
      this.$element[0].removeEventListener("mousemove", this.zoom_bind);
      this.$element[0].removeEventListener("mousedown", this.zoom_bind);
      this.$element[0].removeEventListener("mouseup", this.zoom_bind);
      this.zoom_bind = null;
    }
    this.scale = 1;
    this.transform(this.scale, 0, 0);
  }
  IpeyePlayer.prototype.zoomHandler = function(event) {
    event.preventDefault();
    if (event.type == "wheel") {

      var delta = event.deltaY || event.detail || event.wheelDelta;
      if (delta > 0) {

        if ((this.scale - 0.5) >= 1) {
          this.scale -= 0.5;
        } else {
          this.scale = 1;
        }


      } else {
        if ((this.scale + 0.5) <= 10) {
          this.scale += 0.5;
        } else {
          this.scale = 10;
        }

      }


      this.transform(event.offsetX, event.offsetY);
      //  }

    } else if (event.type == 'mousemove') {
      if (this.mousedown) {
        var max_x=(this.$element.width()/2)*(this.scale-1);
        var max_y=(this.$element.height()/2)*(this.scale-1);
        var x=this.translate_x+(event.offsetX - this.down_pos_x);
        var y=this.translate_y+(event.offsetY - this.down_pos_y);
        if(x>=max_x){
          x=max_x;
        }
        if(x<=(-1*max_x)){
          x=(-1*max_x);
        }
        if(y>=max_y){
          y=max_y;
        }
        if(y<=(-1*max_y)){
          y=(-1*max_y);
        }

        $(this.video).css('transform', 'translate(' + x + 'px, ' + y + 'px) scale(' + this.scale + ')');
        this.translate_x=x;
        this.translate_y=y;

      }
    } else if (event.type == 'mousedown') {
      var matrix = new WebKitCSSMatrix($(this.video).css('transform'));
      this.mousedown = true;
      this.down_pos_x = event.offsetX;
      this.down_pos_y = event.offsetY;

      this.translate_x=matrix.m41;
      this.translate_y=matrix.m42;

    } else if (event.type == 'mouseup') {
      this.mousedown = false;
    }

  }

  IpeyePlayer.prototype.transform = function(pos_x, pos_y) {

    var x = pos_x * this.scale || 0;
    var y = pos_y * this.scale || 0;

    x = ($(this.video).width() * this.scale / 2 - $(this.video).width() / 2) - pos_x * (this.scale - 1);
    y = ($(this.video).height() * this.scale / 2 - $(this.video).height() / 2) - pos_y * (this.scale - 1);


    $(this.video).css('transform', 'translate(' + x + 'px, ' + y + 'px) scale(' + this.scale + ')');
    $(this.video).css('transform-origin',  'center');
  }

  IpeyePlayer.prototype.restartStream = function() {
    var _this = this;
    if (!!_this.ws) {

      _this.websocketRemoveListeners();
      _this.ws.close(1000);
    }
    if(this.webrtc==null){
      this.websocketConnect();
    }

  }

  IpeyePlayer.prototype.screenshot = function() {

    this.canvas_screenshot.width = this.video.videoWidth;
    this.canvas_screenshot.height = this.video.videoHeight;

    this.ctx_screenshot.drawImage(this.video, 0, 0, this.canvas_screenshot.width, this.canvas_screenshot.height);

    this.canvas_screenshot.toBlob(function(blob) {
      // console.log(blob);
      saveAs(blob, "screen.png");
    }, "image/png");
  }
  //Если вулючен debug то будет писать в консоль о действиях
  IpeyePlayer.prototype.logging = function(message) {
    if (this.options.debug) {
      console.log(message);
    }
  }

  IpeyePlayer.prototype.blurplayer = function(message) {
    this.video.muted = true;
    $(this.video).css("filter", "blur(10px) grayscale(1)");
    this.playerShowMessage(message);
  }

  IpeyePlayer.prototype.destroy = function() {

    //console.log('destroy & create new f');

    if (!!this.video) {
      this.video.pause();
      this.video.removeAttribute('src'); // empty source
      this.video.load();
    }
    if(this.webrtc!=null){
      clearInterval(this.webrtcSendChannelInterval);
      this.webrtc.close();
      this.video.srcObject = null;
      this.webrtc=null;
    }
    if (!!this.ws) {
      this.ws.onerror = null;
      this.ws.onopen = null;
      this.ws.onmessage = null;
      this.ws.onclose = null;
      this.ws.close(1000);
    }
    if (!!this.setTimeoutId) {
      clearInterval(this.setTimeoutId);
    }
    if (!!this.messageTimeout) {
      clearInterval(this.messageTimeout);
    }
    this.$element.empty();
    this.$element.unbind();
    if (!!this.window_bind) {
      window.removeEventListener('resize', this.window_bind);
    }
    this.disableZoom();

  }


  IpeyePlayer.prototype.Utf8ArrayToStr = function(array) {
    var out, i, len, c;
    var char2, char3;

    out = "";
    len = array.length;
    i = 0;
    while (i < len) {
      c = array[i++];
      switch (c >> 4) {
        case 0:
        case 1:
        case 2:
        case 3:
        case 4:
        case 5:
        case 6:
        case 7:
          // 0xxxxxxx
          out += String.fromCharCode(c);
          break;
        case 12:
        case 13:
          // 110x xxxx   10xx xxxx
          char2 = array[i++];
          out += String.fromCharCode(((c & 0x1F) << 6) | (char2 & 0x3F));
          break;
        case 14:
          // 1110 xxxx  10xx xxxx  10xx xxxx
          char2 = array[i++];
          char3 = array[i++];
          out += String.fromCharCode(((c & 0x0F) << 12) |
            ((char2 & 0x3F) << 6) |
            ((char3 & 0x3F) << 0));
          break;
      }
    }

    return out;
  }

  IpeyePlayer.prototype.browserDetect = function() {
    var Browser;
    var ua = self.navigator.userAgent.toLowerCase();
    var match =
      /(edge)\/([\w.]+)/.exec(ua) ||
      /(opr)[\/]([\w.]+)/.exec(ua) ||
      /(chrome)[ \/]([\w.]+)/.exec(ua) ||
      /(iemobile)[\/]([\w.]+)/.exec(ua) ||
      /(version)(applewebkit)[ \/]([\w.]+).*(safari)[ \/]([\w.]+)/.exec(ua) ||
      /(webkit)[ \/]([\w.]+).*(version)[ \/]([\w.]+).*(safari)[ \/]([\w.]+)/.exec(
        ua
      ) ||
      /(webkit)[ \/]([\w.]+)/.exec(ua) ||
      /(opera)(?:.*version|)[ \/]([\w.]+)/.exec(ua) ||
      /(msie) ([\w.]+)/.exec(ua) ||
      (ua.indexOf("trident") >= 0 && /(rv)(?::| )([\w.]+)/.exec(ua)) ||
      (ua.indexOf("compatible") < 0 && /(firefox)[ \/]([\w.]+)/.exec(ua)) || [];
    var platform_match =
      /(ipad)/.exec(ua) ||
      /(ipod)/.exec(ua) ||
      /(windows phone)/.exec(ua) ||
      /(iphone)/.exec(ua) ||
      /(kindle)/.exec(ua) ||
      /(android)/.exec(ua) ||
      /(windows)/.exec(ua) ||
      /(mac)/.exec(ua) ||
      /(linux)/.exec(ua) ||
      /(cros)/.exec(ua) || [];
    var matched = {
      browser: match[5] || match[3] || match[1] || "",
      version: match[2] || match[4] || "0",
      majorVersion: match[4] || match[2] || "0",
      platform: platform_match[0] || ""
    };
    var browser = {};

    if (matched.browser) {
      browser[matched.browser] = true;
      var versionArray = matched.majorVersion.split(".");
      browser.version = {
        major: parseInt(matched.majorVersion, 10),
        string: matched.version
      };

      if (versionArray.length > 1) {
        browser.version.minor = parseInt(versionArray[1], 10);
      }

      if (versionArray.length > 2) {
        browser.version.build = parseInt(versionArray[2], 10);
      }
    }

    if (matched.platform) {
      browser[matched.platform] = true;
    }

    if (browser.chrome || browser.opr || browser.safari) {
      browser.webkit = true;
    } // MSIE. IE11 has 'rv' identifer

    if (browser.rv || browser.iemobile) {
      if (browser.rv) {
        delete browser.rv;
      }

      var msie = "msie";
      matched.browser = msie;
      browser[msie] = true;
    } // Microsoft Edge

    if (browser.edge) {
      delete browser.edge;
      var msedge = "msedge";
      matched.browser = msedge;
      browser[msedge] = true;
    } // Opera 15+

    if (browser.opr) {
      var opera = "opera";
      matched.browser = opera;
      browser[opera] = true;
    } // Stock android browsers are marked as Safari

    if (browser.safari && browser.android) {
      var android = "android";
      matched.browser = android;
      browser[android] = true;
    }

    browser.name = matched.browser;
    browser.platform = matched.platform;


    return browser;
  }

  IpeyePlayer.prototype.openFullscreen = function(elem) {
    this.$element.css('height', '100%');
    if (elem.requestFullscreen) {
      elem.requestFullscreen();
    } else if (elem.mozRequestFullScreen) {
      /* Firefox */
      elem.mozRequestFullScreen();
    } else if (elem.webkitRequestFullscreen) {
      /* Chrome, Safari and Opera */
      elem.webkitRequestFullscreen();
    } else if (elem.msRequestFullscreen) {
      /* IE/Edge */
      elem.msRequestFullscreen();
    }

  }

  IpeyePlayer.prototype.closeFullscreen = function() {
    this.$element.css('height', '');
    if (document.exitFullscreen) {
      document.exitFullscreen();
    } else if (document.mozCancelFullScreen) {
      /* Firefox */
      document.mozCancelFullScreen();
    } else if (document.webkitExitFullscreen) {
      /* Chrome, Safari and Opera */
      document.webkitExitFullscreen();
    } else if (document.msExitFullscreen) {
      /* IE/Edge */
      document.msExitFullscreen();
    }
  }

  IpeyePlayer.prototype.fullscreenEnabled = function() {
    return !!(
      document.fullscreenEnabled ||
      document.webkitFullscreenEnabled ||
      document.mozFullScreenEnabled ||
      document.msFullscreenEnabled
    )
  }

  IpeyePlayer.prototype.checkFlash = function() {
    var flashinstalled = false;
    if (navigator.plugins) {
      if (navigator.plugins["Shockwave Flash"]) {
        flashinstalled = true;
      } else if (navigator.plugins["Shockwave Flash 2.0"]) {
        flashinstalled = true;
      }
    } else if (navigator.mimeTypes) {
      var x = navigator.mimeTypes['application/x-shockwave-flash'];
      if (x && x.enabledPlugin) {
        flashinstalled = true;
      }
    } else {
      // на всякий случай возвращаем true в случае некоторых экзотических браузеров
      flashinstalled = true;
    }
    return flashinstalled;
  }
  IpeyePlayer.prototype.canvasReposition = function() {
    if (!!this.video && !!this.video.videoWidth) {
      if (this.$element.width() / this.$element.height() == this.video.videoWidth / this.video.videoHeight) {
        this.canvas.width = this.$element.width();
        this.canvas.height = this.$element.height();
        this.$element.find('canvas.draw_canvas').css({
          'top': 0,
          'left': 0
        })
      } else if (this.$element.width() / this.$element.height() < this.video.videoWidth / this.video.videoHeight) {
        this.canvas.width = this.$element.width();
        this.canvas.height = Math.ceil((this.$element.width() * this.video.videoHeight) / this.video.videoWidth);
        $(this.canvas).css({
          'top': Math.ceil((this.$element.height() - this.canvas.height) / 2) + 'px',
          'left': 0
        })
      } else {
        this.canvas.height = this.$element.width();
        this.canvas.width = Math.ceil(this.$element.height() * this.video.videoWidth / this.video.videoHeight);
        $(this.canvas).css({
          'top': 0,
          'left': Math.ceil((this.$element.width() - this.canvas.width) / 2) + 'px'
        })
      }
    }

    if (typeof this.options.oncanvasresize == 'function') {
      this.options.oncanvasresize();
    }
  }

  IpeyePlayer.prototype.webrtcsupport = function (){
    var checksupport = navigator.getUserMedia ||
        navigator.webkitGetUserMedia ||
        navigator.mozGetUserMedia ||
        navigator.msGetUserMedia ||
        window.RTCPeerConnection;
    return checksupport;
  }

  IpeyePlayer.prototype.startWebRtcPlay = function (server,devcode){
    var _this=this;
      this.webrtc=new RTCPeerConnection({
        iceServers: [{
          urls: ["stun:stun.l.google.com:19302"]
        }],
        //sdpSemantics:'plan-b'
      });
      this.webrtc.onnegotiationneeded = async function (e){
        var offer = await _this.webrtc.createOffer();
        await _this.webrtc.setLocalDescription(offer);
      }
      this.webrtc.ontrack = function(event) {
        _this.logging('[webrtc]=>'+event.streams.length + ' track is delivered');
        _this.webrtcMediaStream.addTrack(event.track);
        _this.video.srcObject = _this.webrtcMediaStream;
      }

      this.webrtc.oniceconnectionstatechange = e => _this.logging('[webrtc][oniceconnectionstatechange]=>'+this.webrtc.iceConnectionState);
      /*******************************************************************************************************/
      this.webrtc.onicecandidate = async function(event) {
        if(event.candidate) return;

        var url ="//"+server+"/webrtc/"+devcode+"/channel/0/webrtc?uuid=test&channel=0";


        $.post(url, {

          data: btoa(_this.webrtc.localDescription.sdp)
        }, function(data) {

          try {
            _this.webrtc.setRemoteDescription(new RTCSessionDescription({
              type: 'answer',
              sdp: atob(data)
            }))
          } catch (e) {
            console.warn(e);
          }

        });


      }
      this.webrtc.onicecandidateerror = function(event) {

      }
      this.webrtc.onicegatheringstatechange = function(event) {

      }
      this.webrtc.onsignalingstatechange = function(event) {

      }
      /*******************************************************************************************************/
      this.webrtc.addTransceiver('video', {
        'direction': 'sendrecv'
      });



  }

  IpeyePlayer.prototype.svgIcons = function(icon) {
    var svg = '';
    switch (icon) {
      case 'zoom':
        svg = '<svg version="1.1" id="Capa_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"	 viewBox="0 0 260 260" ><path d="M258.932,129.176c-2.729-16.26-10.822-24.907-17.13-29.299c-6.565-4.57-13.466-6.279-18.713-6.848 c-6.651-8.935-16.266-13.82-27.325-13.82c-2.416,0-4.879,0.231-7.353,0.689c-6.555-7.522-15.633-11.626-25.804-11.626	c-0.188,0-0.607,0.002-0.607,0.005V30.555c0-16.232-12.269-28.473-28.5-28.473S105,14.323,105,30.555v105.355 c0,1.642-1.427,2.193-2.111,2.193c-0.329,0-0.622-0.088-0.968-0.27c-9.366-4.95-19.949-9.698-33.34-14.941	c-3.819-1.495-7.712-2.254-11.586-2.254c-12.74-0.002-23.695,8.157-27.265,20.302c-3.646,12.398,1.744,25.393,13.727,33.105		c7.606,4.895,15.615,8.834,23.36,12.644c11.766,5.786,22.879,11.252,27.29,17.764c3.83,5.654,12.267,21.683,22.569,42.598		c3.268,6.634,9.893,10.921,17.292,10.921h76.996c9.627,0,17.844-7.363,19.111-16.906l3.916-29.559C255.867,177.947,263.096,153.98,258.932,129.176z M223.378,205.728c-0.601,0.917-0.994,1.959-1.139,3.046l-4.059,30.711	c-0.48,3.617-3.566,6.486-7.216,6.486h-76.996c-2.772,0-5.302-1.736-6.526-4.224c-4.868-9.884-17.427-35.126-23.398-43.941		c-9.66-14.26-34.499-21.201-54.09-33.81c-17.068-10.985-8.381-31.34,7.049-31.338c2.28,0,4.709,0.434,7.221,1.418		c11.25,4.404,22.072,9.058,32.127,14.371c2.156,1.14,4.34,1.659,6.539,1.659c7.423,0,14.111-5.959,14.111-14.196V30.555		c0-10.982,8.264-16.473,16.5-16.473s16.5,5.49,16.5,16.473v42.839c0,4.095,3.385,7.271,7.348,7.271c0.357,0,0.743-0.026,1.11-0.08		c1.304-0.19,2.717-0.312,4.181-0.312c5.899,0,12.786,1.962,18.075,9.209c1.382,1.893,3.576,2.929,5.853,2.929		c0.627,0,1.261-0.078,1.885-0.239c2.043-0.524,4.568-0.962,7.311-0.962c6.124,0,13.339,2.18,18.706,10.432		c1.24,1.907,3.356,3.06,5.628,3.182c7.362,0.394,23.233,3.9,27,26.341C250.668,152.43,244.602,173.384,223.378,205.728z"/>	<path d="M48.333,99.971c0-3.313-2.687-6-6-6H21.434L75,39.96v20.9c0,3.313,2.687,6,6,6s6-2.687,6-6v-36		c0-3.313-2.354-5.889-5.667-5.889h-35c-3.313,0-6,2.686-6,6s2.687,6,6,6h21.128L12,85.99V64.86c0-3.313-2.686-6-6-6s-6,2.687-6,6		v35c0,3.313,3.02,6.111,6.333,6.111h36C45.646,105.971,48.333,103.285,48.333,99.971z"/></svg>';
        break;
      case 'speaker':
        svg = '<svg version="1.1"  xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px" viewBox="0 0 52.026 52.026" xml:space="preserve"><path d="M28.404,3.413c-0.976-0.552-2.131-0.534-3.09,0.044c-0.046,0.027-0.09,0.059-0.13,0.093L11.634,15.013H1 c-0.553,0-1,0.447-1,1v19c0,0.266,0.105,0.52,0.293,0.707S0.734,36.013,1,36.013l10.61-0.005l13.543,12.44		c0.05,0.046,0.104,0.086,0.161,0.12c0.492,0.297,1.037,0.446,1.582,0.446c0.517-0.001,1.033-0.134,1.508-0.402		C29.403,48.048,30,47.018,30,45.857V6.169C30,5.008,29.403,3.978,28.404,3.413z M28,45.857c0,0.431-0.217,0.81-0.579,1.015		c-0.155,0.087-0.548,0.255-1,0.026L13,34.569v-4.556c0-0.553-0.447-1-1-1s-1,0.447-1,1v3.996l-9,0.004v-17h9v4c0,0.553,0.447,1,1,1		s1-0.447,1-1v-4.536l13.405-11.34c0.461-0.242,0.86-0.07,1.016,0.018C27.783,5.36,28,5.739,28,6.169V45.857z"/>	<path d="M38.797,7.066c-0.523-0.177-1.091,0.103-1.269,0.626c-0.177,0.522,0.103,1.091,0.626,1.269		c7.101,2.411,11.872,9.063,11.872,16.553c0,7.483-4.762,14.136-11.849,16.554c-0.522,0.178-0.802,0.746-0.623,1.27		c0.142,0.415,0.53,0.677,0.946,0.677c0.107,0,0.216-0.017,0.323-0.054c7.896-2.693,13.202-10.106,13.202-18.446		C52.026,17.166,46.71,9.753,38.797,7.066z" />	<path d="M43.026,25.513c0-5.972-4.009-11.302-9.749-12.962c-0.533-0.151-1.084,0.152-1.238,0.684		c-0.153,0.53,0.152,1.085,0.684,1.238c4.889,1.413,8.304,5.953,8.304,11.04s-3.415,9.627-8.304,11.04		c-0.531,0.153-0.837,0.708-0.684,1.238c0.127,0.438,0.526,0.723,0.961,0.723c0.092,0,0.185-0.013,0.277-0.039		C39.018,36.815,43.026,31.485,43.026,25.513z"/></svg>';
        break;
      case 'muted':
        svg = '<svg version="1.1"  xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px"	 viewBox="0 0 54 54" style="enable-background:new 0 0 54 54;" xml:space="preserve"><path d="M46.414,26l7.293-7.293c0.391-0.391,0.391-1.023,0-1.414s-1.023-0.391-1.414,0L45,24.586l-7.293-7.293		c-0.391-0.391-1.023-0.391-1.414,0s-0.391,1.023,0,1.414L43.586,26l-7.293,7.293c-0.391,0.391-0.391,1.023,0,1.414		C36.488,34.902,36.744,35,37,35s0.512-0.098,0.707-0.293L45,27.414l7.293,7.293C52.488,34.902,52.744,35,53,35		s0.512-0.098,0.707-0.293c0.391-0.391,0.391-1.023,0-1.414L46.414,26z"/>	<path d="M28.404,4.4c-0.975-0.552-2.131-0.534-3.09,0.044c-0.046,0.027-0.09,0.059-0.13,0.093L11.634,16H1c-0.553,0-1,0.447-1,1v19		c0,0.266,0.105,0.52,0.293,0.707S0.734,37,1,37l10.61-0.005l13.543,12.44c0.05,0.046,0.104,0.086,0.161,0.12		c0.492,0.297,1.037,0.446,1.582,0.446c0.517-0.001,1.033-0.134,1.508-0.402C29.403,49.035,30,48.005,30,46.844V7.156		C30,5.995,29.403,4.965,28.404,4.4z M28,46.844c0,0.431-0.217,0.81-0.579,1.015c-0.155,0.087-0.548,0.255-1,0.026L13,35.556V31		c0-0.553-0.447-1-1-1s-1,0.447-1,1v3.996L2,35V18h9v4c0,0.553,0.447,1,1,1s1-0.447,1-1v-4.536l13.405-11.34	c0.46-0.242,0.86-0.07,1.016,0.018C27.783,6.347,28,6.725,28,7.156V46.844z"/></svg>';
        break;
      case 'expand':
        svg = '<svg version="1.1"  xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px" viewBox="0 0 513.32 513.32" style="enable-background:new 0 0 513.32 513.32;" xml:space="preserve">			<polygon points="177.523,305.853 42.777,440.6 42.777,342.213 0,342.213 0,513.32 171.107,513.32 171.107,470.543 				74.859,470.543 209.605,335.797 			"/>			<polygon points="470.543,440.6 72.72,42.777 171.107,42.777 171.107,0 0,0 0,171.107 42.777,171.107 42.777,72.72 440.6,470.543 				342.213,470.543 342.213,513.32 513.32,513.32 513.32,342.213 470.543,342.213 			"/>			<polygon points="342.213,0 342.213,42.777 442.739,42.777 307.992,177.523 337.935,207.467 470.543,74.859 470.543,171.107 				513.32,171.107 513.32,0 			"/></svg>';
        break;
      case 'screenshot':
        svg = '<svg version="1.1"  xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px" viewBox="0 0 1000 1000" style="enable-background:new 0 0 352.456 352.456;" xml:space="preserve"><path d="M743.2,256.8c0,0,98.9-1,98.9,96.9c0,0,9.2,94.8-98.9,96.9c0,0-96.8,2.1-95.8-96.9C647.4,353.7,646.3,266,743.2,256.8z"/><path d="M938,107.9c0,0,52,4.1,52,49.9V842c0,0-3.1,49-52,50V107.9L938,107.9z"/><path d="M10,842.1c0,0,2.1,50,50,50h878V747.3h-95.7L647.4,549.5L500.5,697.2L300.7,450.6L103.9,745.3l-43.8,1L59,157.9H10L10,842.1L10,842.1z"/><path d="M938,107.9H60c0,0-50-1-50,49.9h928L938,107.9L938,107.9z"/></svg>';
        break;
      case "pause":
        svg = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><path fill-rule="evenodd" clip-rule="evenodd" d="M1.712 14.76H6.43V1.24H1.71v13.52zm7.86-13.52v13.52h4.716V1.24H9.573z"></path></svg>';
        break;

      case "play":
        svg = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><path  d="M1.425.35L14.575 8l-13.15 7.65V.35z" ></path></svg>';
        break;
      default:
        svg = '';
        break;
    }
    return svg;
  }


  function init_plugin(options, renewData) {
    return this.each(function() {
      var _this = $(this);
      var data = _this.data('IpeyePlayer');
      if (!data) {
        //ещё не создан - создаем
        _this.data('IpeyePlayer', new IpeyePlayer(_this, options));
      } else {
        if (typeof options == 'string') {
          //методы контроля

          if (options == 'destroy') {
            _this.data('IpeyePlayer').destroy();
            _this.data('IpeyePlayer', null);
          }
        } else {

          _this.data('IpeyePlayer').destroy();
          _this.data('IpeyePlayer', null);
          _this.data('IpeyePlayer', new IpeyePlayer(_this, options));
        }
      }
    })
  }
  var old = $.fn.IpeyePlayer;
  $.fn.IpeyePlayer = init_plugin;
  $.fn.IpeyePlayer.noConflict = function() {
    $.fn.IpeyePlayer = old;
    return this;
  }

})(jQuery);
/*
 *canvasToblob
 */
(function(view) {
  "use strict";
  var
    Uint8Array = view.Uint8Array,
    HTMLCanvasElement = view.HTMLCanvasElement,
    canvas_proto = HTMLCanvasElement && HTMLCanvasElement.prototype,
    is_base64_regex = /\s*;\s*base64\s*(?:;|$)/i,
    to_data_url = "toDataURL",
    base64_ranks, decode_base64 = function(base64) {
      var
        len = base64.length,
        buffer = new Uint8Array(len / 4 * 3 | 0),
        i = 0,
        outptr = 0,
        last = [0, 0],
        state = 0,
        save = 0,
        rank, code, undef;
      while (len--) {
        code = base64.charCodeAt(i++);
        rank = base64_ranks[code - 43];
        if (rank !== 255 && rank !== undef) {
          last[1] = last[0];
          last[0] = code;
          save = (save << 6) | rank;
          state++;
          if (state === 4) {
            buffer[outptr++] = save >>> 16;
            if (last[1] !== 61 /* padding character */ ) {
              buffer[outptr++] = save >>> 8;
            }
            if (last[0] !== 61 /* padding character */ ) {
              buffer[outptr++] = save;
            }
            state = 0;
          }
        }
      }
      // 2/3 chance there's going to be some null bytes at the end, but that
      // doesn't really matter with most image formats.
      // If it somehow matters for you, truncate the buffer up outptr.
      return buffer;
    };
  if (Uint8Array) {
    base64_ranks = new Uint8Array([
      62, -1, -1, -1, 63, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, -1, -1, -1, 0, -1, -1, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, -1, -1, -1, -1, -1, -1, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51
    ]);
  }
  if (HTMLCanvasElement && (!canvas_proto.toBlob || !canvas_proto.toBlobHD)) {
    if (!canvas_proto.toBlob)
      canvas_proto.toBlob = function(callback, type /*, ...args*/ ) {
        if (!type) {
          type = "image/png";
        }
        if (this.mozGetAsFile) {
          callback(this.mozGetAsFile("canvas", type));
          return;
        }
        if (this.msToBlob && /^\s*image\/png\s*(?:$|;)/i.test(type)) {
          callback(this.msToBlob());
          return;
        }

        var
          args = Array.prototype.slice.call(arguments, 1),
          dataURI = this[to_data_url].apply(this, args),
          header_end = dataURI.indexOf(","),
          data = dataURI.substring(header_end + 1),
          is_base64 = is_base64_regex.test(dataURI.substring(0, header_end)),
          blob;
        if (Blob.fake) {
          // no reason to decode a data: URI that's just going to become a data URI again
          blob = new Blob
          if (is_base64) {
            blob.encoding = "base64";
          } else {
            blob.encoding = "URI";
          }
          blob.data = data;
          blob.size = data.length;
        } else if (Uint8Array) {
          if (is_base64) {
            blob = new Blob([decode_base64(data)], {
              type: type
            });
          } else {
            blob = new Blob([decodeURIComponent(data)], {
              type: type
            });
          }
        }
        callback(blob);
      };

    if (!canvas_proto.toBlobHD && canvas_proto.toDataURLHD) {
      canvas_proto.toBlobHD = function() {
        to_data_url = "toDataURLHD";
        var blob = this.toBlob();
        to_data_url = "toDataURL";
        return blob;
      }
    } else {
      canvas_proto.toBlobHD = canvas_proto.toBlob;
    }
  }
}(typeof self !== "undefined" && self || typeof window !== "undefined" && window || this.content || this));
/*
 *filesaver
 */
var saveAs = saveAs || (function(view) {
  "use strict";
  // IE <10 is explicitly unsupported
  if (typeof view === "undefined" || typeof navigator !== "undefined" && /MSIE [1-9]\./.test(navigator.userAgent)) {
    return;
  }
  var
    doc = view.document
    // only get URL when necessary in case Blob.js hasn't overridden it yet
    ,
    get_URL = function() {
      return view.URL || view.webkitURL || view;
    },
    save_link = doc.createElementNS("http://www.w3.org/1999/xhtml", "a"),
    can_use_save_link = "download" in save_link,
    click = function(node) {
      var event = new MouseEvent("click");
      node.dispatchEvent(event);
    },
    is_safari = /constructor/i.test(view.HTMLElement) || view.safari,
    is_chrome_ios = /CriOS\/[\d]+/.test(navigator.userAgent),
    throw_outside = function(ex) {
      (view.setImmediate || view.setTimeout)(function() {
        throw ex;
      }, 0);
    },
    force_saveable_type = "application/octet-stream"
    // the Blob API is fundamentally broken as there is no "downloadfinished" event to subscribe to
    ,
    arbitrary_revoke_timeout = 1000 * 40 // in ms
    ,
    revoke = function(file) {
      var revoker = function() {
        if (typeof file === "string") { // file is an object URL
          get_URL().revokeObjectURL(file);
        } else { // file is a File
          file.remove();
        }
      };
      setTimeout(revoker, arbitrary_revoke_timeout);
    },
    dispatch = function(filesaver, event_types, event) {
      event_types = [].concat(event_types);
      var i = event_types.length;
      while (i--) {
        var listener = filesaver["on" + event_types[i]];
        if (typeof listener === "function") {
          try {
            listener.call(filesaver, event || filesaver);
          } catch (ex) {
            throw_outside(ex);
          }
        }
      }
    },
    auto_bom = function(blob) {
      // prepend BOM for UTF-8 XML and text/* types (including HTML)
      // note: your browser will automatically convert UTF-16 U+FEFF to EF BB BF
      if (/^\s*(?:text\/\S*|application\/xml|\S*\/\S*\+xml)\s*;.*charset\s*=\s*utf-8/i.test(blob.type)) {
        return new Blob([String.fromCharCode(0xFEFF), blob], {
          type: blob.type
        });
      }
      return blob;
    },
    FileSaver = function(blob, name, no_auto_bom) {
      if (!no_auto_bom) {
        blob = auto_bom(blob);
      }
      // First try a.download, then web filesystem, then object URLs
      var
        filesaver = this,
        type = blob.type,
        force = type === force_saveable_type,
        object_url, dispatch_all = function() {
          dispatch(filesaver, "writestart progress write writeend".split(" "));
        }
        // on any filesys errors revert to saving with object URLs
        ,
        fs_error = function() {
          if ((is_chrome_ios || (force && is_safari)) && view.FileReader) {
            // Safari doesn't allow downloading of blob urls
            var reader = new FileReader();
            reader.onloadend = function() {
              var url = is_chrome_ios ? reader.result : reader.result.replace(/^data:[^;]*;/, 'data:attachment/file;');
              var popup = view.open(url, '_blank');
              if (!popup) view.location.href = url;
              url = undefined; // release reference before dispatching
              filesaver.readyState = filesaver.DONE;
              dispatch_all();
            };
            reader.readAsDataURL(blob);
            filesaver.readyState = filesaver.INIT;
            return;
          }
          // don't create more object URLs than needed
          if (!object_url) {
            object_url = get_URL().createObjectURL(blob);
          }
          if (force) {
            view.location.href = object_url;
          } else {
            var opened = view.open(object_url, "_blank");
            if (!opened) {
              // Apple does not allow window.open, see https://developer.apple.com/library/safari/documentation/Tools/Conceptual/SafariExtensionGuide/WorkingwithWindowsandTabs/WorkingwithWindowsandTabs.html
              view.location.href = object_url;
            }
          }
          filesaver.readyState = filesaver.DONE;
          dispatch_all();
          revoke(object_url);
        };
      filesaver.readyState = filesaver.INIT;

      if (can_use_save_link) {
        object_url = get_URL().createObjectURL(blob);
        setTimeout(function() {
          save_link.href = object_url;
          save_link.download = name;
          click(save_link);
          dispatch_all();
          revoke(object_url);
          filesaver.readyState = filesaver.DONE;
        });
        return;
      }

      fs_error();
    },
    FS_proto = FileSaver.prototype,
    saveAs = function(blob, name, no_auto_bom) {
      return new FileSaver(blob, name || blob.name || "download", no_auto_bom);
    };
  // IE 10+ (native saveAs)
  if (typeof navigator !== "undefined" && navigator.msSaveOrOpenBlob) {
    return function(blob, name, no_auto_bom) {
      name = name || blob.name || "download";

      if (!no_auto_bom) {
        blob = auto_bom(blob);
      }
      return navigator.msSaveOrOpenBlob(blob, name);
    };
  }

  FS_proto.abort = function() {};
  FS_proto.readyState = FS_proto.INIT = 0;
  FS_proto.WRITING = 1;
  FS_proto.DONE = 2;

  FS_proto.error =
    FS_proto.onwritestart =
    FS_proto.onprogress =
    FS_proto.onwrite =
    FS_proto.onabort =
    FS_proto.onerror =
    FS_proto.onwriteend =
    null;

  return saveAs;
}(
  typeof self !== "undefined" && self ||
  typeof window !== "undefined" && window ||
  this.content
));
// `self` is undefined in Firefox for Android content script context
// while `this` is nsIContentFrameMessageManager
// with an attribute `content` that corresponds to the window

if (typeof module !== "undefined" && module.exports) {
  module.exports.saveAs = saveAs;
} else if ((typeof define !== "undefined" && define !== null) && (define.amd !== null)) {
  define("FileSaver.js", function() {
    return saveAs;
  });
}

function average(nums) {
  return (
    nums.reduce(function(a, b) {
      return a + b;
    }) / nums.length
  );
}
