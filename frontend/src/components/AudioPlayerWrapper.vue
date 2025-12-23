<template>
  <audio ref="audioPlayer" style="display: none;" @canplay="handleCanPlay"></audio>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue';
import { usePlayerStore } from '@/stores/player';
// 引入 Hls
import Hls from 'hls.js';

const audioPlayer = ref(null);
const store = usePlayerStore();
let isReadyToPlay = false;

// 定义 Hls 实例变量
let hls = null;

// 监听要播放的歌曲 URL 变化
watch(() => store.currentSongUrl, (newUrl, oldUrl) => {
  // 清理旧的 hls 实例
  if (hls) {
    hls.destroy();
    hls = null;
  }
  newUrl = location.origin + newUrl; // 补全为完整 URL
  if (newUrl && newUrl !== oldUrl) {
    console.log('Audio source changed to (HLS/Native):', newUrl);
    isReadyToPlay = false; 

    const audio = audioPlayer.value;

    // 检查是否支持 Hls.js
    if (Hls.isSupported()) {
      hls = new Hls();
      hls.loadSource(newUrl);
      hls.attachMedia(audio);
      
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        console.log('HLS manifest loaded, waiting for buffer...');
        // 注意：这里不需要手动调 play，因为下面 handleCanPlay 会处理
        // 或者如果需要自动预加载，可以在这里做逻辑
      });

      hls.on(Hls.Events.ERROR, (event, data) => {
        if (data.fatal) {
          switch (data.type) {
            case Hls.ErrorTypes.NETWORK_ERROR:
              console.error('HLS Network error, trying to recover...');
              hls.startLoad();
              break;
            case Hls.ErrorTypes.MEDIA_ERROR:
              console.error('HLS Media error, trying to recover...');
              hls.recoverMediaError();
              break;
            default:
              console.error('HLS Fatal error, destroying...', data);
              hls.destroy();
              break;
          }
        }
      });

    } else if (audio.canPlayType('application/vnd.apple.mpegurl')) {
      // 针对 Safari 等原生支持 HLS 的浏览器
      audio.src = newUrl;
    } else {
      // 不支持 HLS 的情况，回退到直接设置 src (可能无法播放 m3u8)
      console.warn('HLS not supported and not native. Trying direct source.');
      audio.src = newUrl;
    }

  } else if (!newUrl) {
    audioPlayer.value.src = '';
  }
});

// 当音频可以播放时触发 (原生事件，Hls.js 挂载后同样会触发此事件)
const handleCanPlay = () => {
  console.log('Audio is ready to play (canplay event)');
  isReadyToPlay = true;
  // 如果此时store的状态是播放，就立即开始播放
  if (store.isPlaying) {
    playAudio();
  }
};

const playAudio = () => {
  const playPromise = audioPlayer.value.play();
  if (playPromise !== undefined) {
    playPromise.catch(error => {
      console.error("Audio play failed:", error);
      store.playbackError = error;
    });
  }
};

// 监听播放状态的变化
watch(() => store.isPlaying, (isPlaying) => {
  if (isReadyToPlay) {
    if (isPlaying) {
      playAudio();
    } else {
      audioPlayer.value.pause();
    }
  }
});

// 监听并校准播放进度
watch(() => store.progressMs, (newProgress) => {
  const player = audioPlayer.value;
  if (!player || !isReadyToPlay) return;
  
  // HLS 同样使用 currentTime (秒)
  const timeDifference = Math.abs(newProgress - player.currentTime * 1000);

  if (timeDifference > 2000 && !player.seeking) {
    console.log(`Syncing time: server=${newProgress/1000}s, client=${player.currentTime}s. Correcting...`);
    player.currentTime = newProgress / 1000;
  }
});

watch(() => store.localVolume, (newVolume) => {
  if (audioPlayer.value) {
    audioPlayer.value.volume = newVolume;
  }
});

onMounted(() => {
  if (audioPlayer.value) {
    audioPlayer.value.volume = store.localVolume;
  }
});

// 组件卸载时销毁 HLS 实例
onUnmounted(() => {
  if (hls) {
    hls.destroy();
  }
});

</script>
