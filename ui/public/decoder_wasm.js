/**
 * WebCodec Bridge for H.264 Decoding
 */

class WebCodecBridge {
    constructor() {
        this.decoder = null;
        this.outputFrames = [];
        this.initPromise = null;
        this.isSupported = typeof VideoDecoder !== 'undefined';
        
        // Buffer pool to reuse RGBA buffers and prevent GC pressure
        this.bufferPool = [];
        this.currentWidth = 0;
        this.currentHeight = 0;
        
        if (!this.isSupported) {
            console.error('WebCodec VideoDecoder not supported');
        }
    }

    async initialize(width, height) {
        // Queue initialization to ensure it completes before decoding
        if (!this.initPromise) {
            this.initPromise = this._doInitialize(width, height);
        }
        return this.initPromise;
    }

    async _doInitialize(width, height) {
        if (!this.isSupported) {
            throw new Error('VideoDecoder not supported');
        }

        if (this.decoder) {
            try {
                this.decoder.close();
            } catch (e) {}
        }

        // Clear buffer pool on reinit
        this.bufferPool = [];
        this.currentWidth = width;
        this.currentHeight = height;

        try {
            const config = {
                codec: 'avc1.42001E',
                codedWidth: width,
                codedHeight: height,
            };

            const supported = await VideoDecoder.isConfigSupported(config);
            if (!supported.supported) {
                throw new Error('H.264 codec not supported');
            }

            this.decoder = new VideoDecoder({
                output: (frame) => this._handleFrame(frame),
                error: (error) => this._handleError(error)
            });

            this.decoder.configure(config);
        } catch (error) {
            console.error('VideoDecoder initialization failed:', error);
            throw error;
        }
    }

    async _videoFrameToRGBA(frame) {
        const width = frame.displayWidth;
        const height = frame.displayHeight;
        const size = width * height * 4;
        
        // CRITICAL: Reuse buffer from pool to prevent allocation on every frame
        let rgbaBuffer;
        if (this.bufferPool.length > 0 && this.bufferPool[0].length === size) {
            rgbaBuffer = this.bufferPool.pop();
        } else {
            rgbaBuffer = new Uint8Array(size);
        }
        
        try {
            await frame.copyTo(rgbaBuffer, {
                rect: { x: 0, y: 0, width, height },
                layout: [{ offset: 0, stride: width * 4 }],
                format: 'RGBA'
            });
            return rgbaBuffer;
        } catch (error) {
            // Return buffer to pool even on error
            if (this.bufferPool.length < 3) {
                this.bufferPool.push(rgbaBuffer);
            }
            console.error('copyTo failed:', error);
            throw error;
        }
    }

    _handleFrame(frame) {
        // Critical: Close frame immediately after conversion to prevent accumulation
        this._videoFrameToRGBA(frame)
            .then(rgbaData => {
                // Drop old frames more aggressively - keep only latest 2 frames
                // This prevents memory buildup when Go isn't consuming fast enough
                while (this.outputFrames.length >= 2) {
                    this.outputFrames.shift();
                }
                
                this.outputFrames.push({
                    width: frame.displayWidth,
                    height: frame.displayHeight,
                    data: rgbaData,
                });
            })
            .catch(error => {
                console.error('Frame conversion failed:', error);
            })
            .finally(() => {
                // Close frame exactly once - this is critical for memory
                try {
                    frame.close();
                } catch (e) {
                    // Already closed or invalid
                }
            });
    }

    _handleError(error) {
        console.error('VideoDecoder error:', error);
    }

    close() {
        // Clear queue
        this.outputFrames = [];
        
        // Clear buffer pool
        this.bufferPool = [];
        
        if (this.decoder) {
            try {
                this.decoder.close();
            } catch (e) {}
            this.decoder = null;
        }
        
        this.initPromise = null;
    }

    async decode(encodedData) {
        if (!this.decoder) {
            throw new Error('Decoder not initialized');
        }

        const chunk = new EncodedVideoChunk({
            type: 'key',
            timestamp: 0,
            data: encodedData,
        });
        
        try {
            this.decoder.decode(chunk);
            
            if (this.decoder.decodeQueueSize > 10) {
                await this.decoder.flush();
            }
        } catch (error) {
            console.error('Decode error:', error);

            throw error;
        }
    }

    getNextFrame() {
        if (this.outputFrames.length > 0) {
            const frame = this.outputFrames.shift();
            
            // Return the previous buffer to pool for reuse
            // Go has copied the data, so we can reuse this buffer
            if (frame?.data && this.bufferPool.length < 3) {
                this.bufferPool.push(frame.data);
            }
            
            return frame;
        }
        return null;
    }

}

// Create global instance
const webCodecBridge = new WebCodecBridge();

// Initialize queue for async initialization
let initQueue = Promise.resolve();

window.wasmInitializeDecoder = function(width, height) {
    // Queue initialization without blocking Go
    initQueue = initQueue
        .then(() => webCodecBridge.initialize(width, height))
        .catch(error => {
            console.error('Decoder init failed:', error);
        });
    return 0;
};

window.wasmDecodeH264 = function(data, length) {
    if (!webCodecBridge.decoder) {
        console.warn('Decoder not ready, skipping frame');
        return 1;
    }

    try {
        webCodecBridge.decode(data.subarray(0, length))
            .catch(error => {
                console.error('Decode error:', error);
            });
        return 0;
    } catch (error) {
        console.error('Decode failed:', error);
        return 1;
    }
};

window.wasmGetNextFrame = function() {
    return webCodecBridge.getNextFrame();
};

window.wasmCloseDecoder = function() {
    webCodecBridge.close();
};
