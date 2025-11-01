/**
 * WebCodec Bridge for H.264 Decoding
 * 
 * This module provides a bridge between Go WASM and the browser's WebCodec API
 * to decode H.264 video buffers into RGBA frames for rendering.
 * 
 * The WebCodec API (VideoDecoder) offers hardware-accelerated video decoding
 * directly in the browser, eliminating the need for FFmpeg.
 */

class WebCodecBridge {
    constructor() {
        this.decoder = null;
        this.outputFrames = [];
        this.outputFramesMutex = false;
        this.width = 0;
        this.height = 0;
        this.isInitialized = false;
        this.initPromise = null;
        this.initError = null;

        // Check browser support
        if (typeof VideoDecoder === 'undefined') {
            console.error('WebCodec VideoDecoder is not supported in this browser');
            this.initError = new Error('WebCodec not supported');
        }
    }

    /**
     * Initialize the H.264 decoder with specified dimensions
     * @param {number} width - Video width in pixels
     * @param {number} height - Video height in pixels
     * @returns {Promise<void>}
     */
    async initialize(width, height) {
        if (this.isInitialized) {
            if (this.width === width && this.height === height) {
                return;
            }
            // Different dimensions, need to reinitialize
            this.close();
        }

        this.width = width;
        this.height = height;

        try {
            // Check VideoDecoder support
            const config = {
                codec: 'avc1.42001E', // H.264 Baseline Profile Level 30
                codedWidth: width,
                codedHeight: height,
            };

            const isSupported = await VideoDecoder.isConfigSupported(config);
            if (!isSupported.supported) {
                throw new Error('H.264 codec not supported: ' + JSON.stringify(config));
            }

            this.decoder = new VideoDecoder({
                output: this._handleDecodedFrame.bind(this),
                error: this._handleDecoderError.bind(this),
            });

            this.decoder.configure(config);
            this.isInitialized = true;
            console.log(`WebCodec decoder initialized: ${width}x${height}`);
        } catch (error) {
            this.initError = error;
            this.isInitialized = false;
            throw error;
        }
    }

    /**
     * Close and release decoder resources
     */
    close() {
        if (this.decoder) {
            try {
                this.decoder.close();
            } catch (e) {
                console.warn('Error closing decoder:', e);
            }
            this.decoder = null;
        }
        this.outputFrames = [];
        this.isInitialized = false;
    }

    /**
     * Decode a single H.264 encoded chunk
     * @param {Uint8Array} encodedData - H.264 encoded video chunk
     * @param {number} timestamp - Optional timestamp in microseconds
     * @returns {Promise<void>}
     */
    async decode(encodedData, timestamp = 0) {
        if (!this.isInitialized) {
            throw new Error('Decoder not initialized');
        }

        try {
            const chunk = new EncodedVideoChunk({
                type: 'key', // Could be 'key' or 'delta' - we'll treat all as key for safety
                timestamp: timestamp,
                data: new Uint8Array(encodedData),
            });

            this.decoder.decode(chunk);

            // Don't flush every frame - it's slow! Let decoder work async
            // Only flush if queue is getting big
            if (this.outputFrames.length > 3) {
                await this.decoder.flush().catch(() => {
                    // Flush errors are expected if decoder is still processing
                });
            }
        } catch (error) {
            console.error('Decode error:', error);
            throw error;
        }
    }

    /**
     * Get the next decoded frame as RGBA data
     * @returns {Object|null} Frame object with {width, height, data} or null if no frames available
     */
    getNextFrame() {
        // Simple queue without actual mutex (JavaScript is single-threaded)
        if (this.outputFrames.length > 0) {
            return this.outputFrames.shift();
        }
        return null;
    }

    /**
     * Get number of queued frames
     * @returns {number}
     */
    getFrameQueueSize() {
        return this.outputFrames.length;
    }

    /**
     * Clear the output frame queue
     */
    clearFrameQueue() {
        this.outputFrames = [];
    }

    /**
     * Internal callback for decoded frames
     * @private
     */
    _handleDecodedFrame(frame) {
        try {
            // Convert VideoFrame to RGBA asynchronously
            this._videoFrameToRGBA(frame).then(rgbaData => {
                this.outputFrames.push({
                    width: frame.displayWidth,
                    height: frame.displayHeight,
                    data: rgbaData,
                });

                // Keep only recent frames to avoid memory buildup
                if (this.outputFrames.length > 5) {
                    this.outputFrames.shift();
                }
            }).catch(error => {
                console.error('Error converting frame to RGBA:', error);
            }).finally(() => {
                frame.close();
            });
        } catch (error) {
            console.error('Error processing decoded frame:', error);
            frame.close();
        }
    }

    /**
     * Internal error handler for decoder
     * @private
     */
    _handleDecoderError(error) {
        console.error('VideoDecoder error:', error);
    }

    /**
     * Convert a VideoFrame to RGBA Uint8Array
     * Uses VideoFrame.copyTo() for maximum performance
     * @private
     */
    async _videoFrameToRGBA(frame) {
        const width = frame.displayWidth;
        const height = frame.displayHeight;
        const pixelCount = width * height;
        const rgbaSize = pixelCount * 4;

        try {
            // Try using copyTo for fastest conversion (direct pixel access)
            if (typeof frame.copyTo === 'function') {
                return await this._copyToRGBA(frame, width, height);
            }

            // Fallback to Canvas method
            return await this._canvasToRGBA(frame, width, height);
        } catch (error) {
            console.error('Error converting frame to RGBA:', error);
            // Return blank frame
            return new Uint8Array(rgbaSize);
        }
    }

    /**
     * Fast path: Use VideoFrame.copyTo() for direct pixel access
     * @private
     */
    async _copyToRGBA(frame, width, height) {
        const pixelCount = width * height;
        
        // Try to copy as RGBA directly
        try {
            const rgbaBuffer = new Uint8Array(pixelCount * 4);
            
            await frame.copyTo(rgbaBuffer, {
                rect: { x: 0, y: 0, width: width, height: height },
                layout: [{ offset: 0, stride: width * 4 }],
                format: 'RGBA'
            });
            
            return rgbaBuffer;
        } catch (e) {
            // If RGBA fails, try reading the raw format and converting
            console.debug('RGBA copyTo failed, falling back to format conversion:', e.message);
            
            // Try copying raw YUV data and converting
            const layout = frame.allocationSize('RGBA');
            const buffer = new ArrayBuffer(layout);
            
            try {
                await frame.copyTo(new Uint8Array(buffer), {
                    layout: [{ offset: 0, stride: width * 4 }],
                    format: 'RGBA'
                });
                return new Uint8Array(buffer);
            } catch (e2) {
                console.debug('Format conversion failed:', e2.message);
                throw e;  // Re-throw original error
            }
        }
    }

    /**
     * Fallback: Use Canvas for frame conversion (slower)
     * @private
     */
    async _canvasToRGBA(frame, width, height) {
        try {
            // Use OffscreenCanvas if available (better performance)
            if (typeof OffscreenCanvas !== 'undefined') {
                const canvas = new OffscreenCanvas(width, height);
                const ctx = canvas.getContext('2d', { willReadFrequently: true });

                if (ctx === null) {
                    throw new Error('Failed to get 2D context');
                }

                const bitmap = await createImageBitmap(frame);
                ctx.drawImage(bitmap, 0, 0);
                bitmap.close();

                const imageData = ctx.getImageData(0, 0, width, height);
                return new Uint8Array(imageData.data);
            } else {
                // Fallback to regular Canvas
                const canvas = document.createElement('canvas');
                canvas.width = width;
                canvas.height = height;
                const ctx = canvas.getContext('2d', { willReadFrequently: true });

                if (ctx === null) {
                    throw new Error('Failed to get 2D context');
                }

                const bitmap = await createImageBitmap(frame);
                ctx.drawImage(bitmap, 0, 0);
                bitmap.close();

                const imageData = ctx.getImageData(0, 0, width, height);
                return new Uint8Array(imageData.data);
            }
        } catch (error) {
            console.error('Canvas conversion failed:', error);
            throw error;
        }
    }

    /**
     * Get current decoder state
     * @returns {Object}
     */
    getState() {
        return {
            isInitialized: this.isInitialized,
            width: this.width,
            height: this.height,
            decoderState: this.decoder ? this.decoder.state : 'closed',
            queuedFrames: this.outputFrames.length,
            lastError: this.initError ? this.initError.message : null,
        };
    }
}

// Create global instance
const webCodecBridge = new WebCodecBridge();
console.log('[WebCodecBridge] Instantiated:', webCodecBridge);
console.log('[WebCodecBridge] VideoDecoder available:', typeof VideoDecoder !== 'undefined');

/**
 * Global functions exposed to Go WASM
 */

// Queue for pending async initialization tasks
let initializationQueue = Promise.resolve();

/**
 * Initialize WebCodec decoder - called from Go
 * Queue initialization and return immediately (async operations happen in background)
 * @param {number} width
 * @param {number} height
 * @returns {number} 0 (always succeeds immediately, async init happens in background)
 */
window.wasmInitializeDecoder = function(width, height) {
    // Queue the initialization to happen asynchronously
    initializationQueue = initializationQueue
        .then(() => webCodecBridge.initialize(width, height))
        .catch((error) => {
            console.error('WASM decoder initialization failed:', error);
            webCodecBridge.initError = error;
        });
    
    // Return immediately for synchronous Go code
    return 0;
};

/**
 * Decode H.264 buffer - called from Go (synchronous wrapper for async decode)
 * @param {Uint8Array} data - H.264 encoded data
 * @param {number} length - Length of data
 * @param {number} timestamp - Timestamp in microseconds
 * @returns {number} 0 on success, 1 on error
 */
window.wasmDecodeH264 = function(data, length, timestamp = 0) {
    // Ensure decoder is ready before decoding
    if (!webCodecBridge.isInitialized) {
        if (webCodecBridge.initError) {
            console.error('Decoder not ready:', webCodecBridge.initError);
            return 1;
        }
        // Still initializing
        console.warn('Decoder still initializing, skipping frame');
        return 1;
    }
    
    try {
        if (length > 0 && data && data.length >= length) {
            // Submit decode asynchronously (don't await in Go code)
            webCodecBridge.decode(data.subarray(0, length), timestamp)
                .catch((error) => {
                    console.error('WASM H.264 decode failed:', error);
                });
        }
        return 0;
    } catch (error) {
        console.error('WASM H.264 decode failed:', error);
        return 1;
    }
};

/**
 * Get next decoded frame - called from Go
 * @returns {Object|null} Frame with {width, height, data: Uint8Array} or null
 */
window.wasmGetNextFrame = function() {
    return webCodecBridge.getNextFrame();
};

/**
 * Get number of queued frames - called from Go
 * @returns {number}
 */
window.wasmGetFrameQueueSize = function() {
    return webCodecBridge.getFrameQueueSize();
};

/**
 * Clear frame queue - called from Go
 */
window.wasmClearFrameQueue = function() {
    webCodecBridge.clearFrameQueue();
};

/**
 * Close decoder - called from Go
 */
window.wasmCloseDecoder = function() {
    webCodecBridge.close();
};

/**
 * Get decoder state - for debugging
 */
window.wasmGetDecoderState = function() {
    return JSON.stringify(webCodecBridge.getState());
};

/**
 * Check if the WebCodec bridge is fully initialized and ready
 * @returns {number} 1 if ready, 0 if still initializing, -1 if error
 */
window.wasmIsBridgeReady = function() {
    if (webCodecBridge.initError) {
        return -1;
    }
    return webCodecBridge.isInitialized ? 1 : 0;
};

console.log('WebCodec bridge loaded successfully');
