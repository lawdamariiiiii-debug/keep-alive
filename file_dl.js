const _rootDomain = location.hostname.split('.').slice(-2).join('.');
const DOWNLOAD_SERVICE_URL = location.origin;
const CDN_URLS = [
    'https://cache1.filester.me',
    'https://cache6.filester.me',
    'https://cache00.filester.me'
];

function getCdnUrl() {
    return CDN_URLS[Math.floor(Math.random() * CDN_URLS.length)];
}

function getFileSlugFromURL() {
    const pathParts = window.location.pathname.split('/');
    return pathParts[pathParts.length - 1];
}

async function generatePublicDownloadToken() {
    const fileSlug = getFileSlugFromURL();
    
    try {
        const response = await fetch(`${DOWNLOAD_SERVICE_URL}/api/public/download`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                file_slug: fileSlug
            })
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(errorText || `HTTP error! status: ${response.status}`);
        }

        return await response.json();
    } catch (error) {
        console.error('Error generating download token:', error);
        throw error;
    }
}

async function generatePublicViewToken() {
    const fileSlug = getFileSlugFromURL();
    
    try {
        const response = await fetch(`${DOWNLOAD_SERVICE_URL}/api/public/view`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                file_slug: fileSlug
            })
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(errorText || `HTTP error! status: ${response.status}`);
        }

        return await response.json();
    } catch (error) {
        console.error('Error generating view token:', error);
        throw error;
    }
}

async function handleDownload(event) {
    if (event) {
        event.preventDefault();
        event.stopPropagation();
    }
    
    const downloadButton = document.getElementById('downloadButton');
    if (!downloadButton) return;
    
    const originalText = downloadButton.innerHTML;
    
    downloadButton.innerHTML = `
        <svg class="download-icon" style="width: 16px; height: 16px; animation: spin 1s linear infinite;" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17V6h2v8.17l3.59-3.58L17 12l-5 5z"/>
        </svg>
        Generating download link...
    `;
    downloadButton.style.pointerEvents = 'none';
    
    try {
        const tokenData = await generatePublicDownloadToken();
        const downloadUrl = `${getCdnUrl()}${tokenData.download_url}?download=true`;
        
        const newWindow = window.open(downloadUrl, '_blank');
        
        if (!newWindow || newWindow.closed || typeof newWindow.closed === 'undefined') {
            window.location.href = downloadUrl;
        }
        
        setTimeout(() => {
            downloadButton.innerHTML = originalText;
            downloadButton.style.pointerEvents = 'auto';
        }, 3000);
        
    } catch (error) {
        console.error('Download error:', error);
        
        downloadButton.innerHTML = `
            <svg class="download-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
            </svg>
            Download failed - Try again
        `;
        
        setTimeout(() => {
            downloadButton.innerHTML = originalText;
            downloadButton.style.pointerEvents = 'auto';
        }, 3000);
    }
}

async function setupImage() {
    const imageElement = document.getElementById('imagePreview');
    if (!imageElement) return;
    
    try {
        const tokenData = await generatePublicViewToken();
        const imageUrl = `${getCdnUrl()}${tokenData.view_url}`;
        
        imageElement.style.opacity = '0';
        
        imageElement.onload = function() {
            imageElement.style.opacity = '1';
            
            const container = imageElement.closest('.image-container');
            if (container) {
                container.addEventListener('click', function(e) {
                    if (e.target === imageElement || e.target === container) {
                        container.classList.toggle('fullscreen');
                        
                        if (container.classList.contains('fullscreen')) {
                            const escHandler = function(event) {
                                if (event.key === 'Escape') {
                                    container.classList.remove('fullscreen');
                                    document.removeEventListener('keydown', escHandler);
                                }
                            };
                            document.addEventListener('keydown', escHandler);
                        }
                    }
                });
            }
        };
        
        imageElement.onerror = function() {
            const container = document.querySelector('.image-container');
            if (container) {
                container.innerHTML = `
                    <div class="image-error">
                        <div class="image-error-icon">⚠️</div>
                        <div class="image-error-text">Unable to load image</div>
                    </div>
                `;
            }
        };
        
        imageElement.src = imageUrl;
    } catch (error) {
        console.error('Failed to setup image:', error);
    }
}

async function setupVideo() {
    const videoElement = document.getElementById('videoPlayer');
    if (!videoElement) {
        return;
    }

    const videoContainer = videoElement.closest('.media-container');
    if (!videoContainer) {
        return;
    }

    try {
        const tokenData = await generatePublicViewToken();
        const streamUrl = `${getCdnUrl()}${tokenData.view_url}`;
        
        const sourceTags = videoElement.querySelectorAll('source');
        sourceTags.forEach(source => source.remove());
        
        videoElement.src = streamUrl;
        
        videoElement.load();
        
        const player = new Plyr(videoElement, {
            controls: [
                'play-large',
                'play',
                'progress',
                'current-time',
                'duration',
                'mute',
                'volume',
                'settings',
                'pip',
                'fullscreen'
            ],
            settings: ['captions', 'quality', 'speed', 'loop'],
            speed: { 
                selected: 1, 
                options: [0.5, 0.75, 1, 1.25, 1.5, 1.75, 2] 
            },
            keyboard: { 
                focused: true, 
                global: false 
            },
            tooltips: { 
                controls: true, 
                seek: true 
            },
            fullscreen: { 
                enabled: true, 
                fallback: true, 
                iosNative: true
            },
            storage: { 
                enabled: true, 
                key: 'plyr' 
            },
            playsinline: true,
            clickToPlay: true,
            hideControls: true,
            resetOnEnd: false
        });
        
        // Store player instance globally
        window.videoPlayer = player;

        player.on('error', (event) => {
            console.error('Player error:', event);
        });
        
    } catch (error) {
        console.error('Failed to setup video:', error);
        videoContainer.innerHTML = `
            <div class="media-error">
                <div class="media-error-icon">⚠️</div>
                <div class="media-error-text">Unable to load video</div>
            </div>
        `;
    }
}

async function setupAudio() {
    const audioElement = document.getElementById('audioPlayer');
    if (!audioElement) {
        return;
    }

    try {
        const tokenData = await generatePublicViewToken();
        const audioUrl = `${getCdnUrl()}${tokenData.view_url}`;
        
        const sourceTags = audioElement.querySelectorAll('source');
        sourceTags.forEach(source => source.remove());
        
        audioElement.src = audioUrl;
        
        audioElement.load();
        
        audioElement.addEventListener('loadedmetadata', function() {
            if (typeof Plyr !== 'undefined' && !window.audioPlayer) {
                const player = new Plyr(audioElement, {
                    controls: [
                        'restart',
                        'rewind',
                        'play',
                        'fast-forward',
                        'progress',
                        'current-time',
                        'duration',
                        'mute',
                        'volume',
                        'settings'
                    ],
                    settings: ['speed', 'loop'],
                    speed: { 
                        selected: 1, 
                        options: [0.5, 0.75, 1, 1.25, 1.5, 1.75, 2] 
                    },
                    storage: { 
                        enabled: true, 
                        key: 'plyr_audio' 
                    }
                });

                window.audioPlayer = player;

                player.on('error', (event) => {
                    console.error('Audio player error:', event);
                });
            }
        }, { once: true });
        
        audioElement.addEventListener('error', function(e) {
            console.error('Audio load error:', e);
            const error = audioElement.error;
            if (error) {
                console.error('Error code:', error.code, 'Message:', error.message);
            }
            
            const container = audioElement.closest('.media-container');
            if (container) {
                container.innerHTML = `
                    <div class="media-error">
                        <div class="media-error-icon">⚠️</div>
                        <div class="media-error-text">Unable to load audio</div>
                    </div>
                `;
            }
        });

    } catch (error) {
        console.error('Failed to setup audio:', error);
        const container = document.querySelector('.media-container');
        if (container) {
            container.innerHTML = `
                <div class="media-error">
                    <div class="media-error-icon">⚠️</div>
                    <div class="media-error-text">Unable to load audio: ${error.message}</div>
                </div>
            `;
        }
    }
}

function isImage(mimeType) {
    return mimeType && mimeType.startsWith('image/');
}

function isVideo(mimeType) {
    return mimeType && (mimeType.startsWith('video/') || 
           (mimeType === 'application/octet-stream' && window.fileName && 
            window.fileName.toLowerCase().match(/\.(mp4|webm|mov|avi|mkv)$/)));
}

function isAudio(mimeType) {
    if (mimeType && mimeType.startsWith('audio/')) {
        return true;
    }
    
    if (window.fileName) {
        const audioExtensions = ['.mp3', '.wav', '.ogg', '.m4a', '.aac', '.flac', '.wma'];
        const fileName = window.fileName.toLowerCase();
        return audioExtensions.some(ext => fileName.endsWith(ext));
    }
    
    return false;
}

document.addEventListener('DOMContentLoaded', function() {
    const downloadButton = document.getElementById('downloadButton');
    if (downloadButton) {
        downloadButton.addEventListener('click', handleDownload);
    }

    const fileType = window.fileType;

    if (isImage(fileType)) {
        setupImage();
    } else if (isVideo(fileType)) {
        if (typeof Plyr !== 'undefined') {
            setupVideo();
        } else {
            console.error('Plyr library not loaded - video player unavailable');
        }
    } else if (isAudio(fileType)) {
        if (typeof Plyr !== 'undefined') {
            setupAudio();
        } else {
            console.error('Plyr library not loaded - audio player unavailable');
        }
    }
    
    const style = document.createElement('style');
    style.textContent = `
        @keyframes spin {
            from { transform: rotate(0deg); }
            to { transform: rotate(360deg); }
        }
        
        .video-container {
            min-height: 400px;
            position: relative;
            background: #000;
        }
        
        .media-container {
            position: relative;
            background: #000;
        }
        
        .audio-container {
            min-height: 80px;
            background: transparent;
        }
        
        video {
            width: 100%;
            height: auto;
            min-height: 400px;
            background: #000;
        }
        
        audio {
            width: 100%;
            height: auto;
            min-height: 50px;
        }
        
        .plyr--audio {
            min-height: 50px;
        }
        
        .media-error {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding: 3rem;
            color: #666;
            min-height: 100px;
        }
        
        .media-error-icon {
            font-size: 3rem;
            margin-bottom: 1rem;
            opacity: 0.5;
        }
        
        .media-error-text {
            font-size: 1rem;
            text-align: center;
        }
        
        .image-container.fullscreen {
            position: fixed;
            top: 0;
            left: 0;
            width: 100vw;
            height: 100vh;
            max-height: 100vh;
            z-index: 9999;
            background: rgba(0, 0, 0, 0.95);
            cursor: zoom-out;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        
        .image-container:not(.fullscreen) {
            cursor: zoom-in;
        }
        
        .image-container.fullscreen .image-preview {
            max-width: 95vw;
            max-height: 95vh;
            object-fit: contain;
        }
    `;
    document.head.appendChild(style);
});