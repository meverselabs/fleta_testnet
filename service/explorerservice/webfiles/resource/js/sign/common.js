if (typeof Symbol === "undefined") {
    Symbol = function () {
        function s4() {
            return Math.floor((1 + Math.random()) * 0x10000)
                .toString(16)
                .substring(1);
        }
        return s4() + s4() + '-' + s4() + '-' + s4() + '-' + s4() + '-' + s4() + s4() + s4();
    }
}
if (!Number.MAX_SAFE_INTEGER) {
    Number.MAX_SAFE_INTEGER = Math.pow(2, 53) - 1; // 9007199254740991
}
if (!String.prototype.repeat) {
    String.prototype.repeat = function (count) {
        'use strict';
        if (this == null) {
            throw new TypeError('can\'t convert ' + this + ' to object');
        }
        var str = '' + this;
        count = +count;
        if (count != count) {
            count = 0;
        }
        if (count < 0) {
            throw new RangeError('repeat count must be non-negative');
        }
        if (count == Infinity) {
            throw new RangeError('repeat count must be less than infinity');
        }
        count = Math.floor(count);
        if (str.length == 0 || count == 0) {
            return '';
        }
        // Ensuring count is a 31-bit integer allows us to heavily optimize the
        // main part. But anyway, most current (August 2014) browsers can't handle
        // strings 1 << 28 chars or longer, so:
        if (str.length * count >= 1 << 28) {
            throw new RangeError('repeat count must not overflow maximum string size');
        }
        var maxCount = str.length * count;
        count = Math.floor(Math.log(count) / Math.log(2));
        while (count) {
            str += str;
            count--;
        }
        str += str.substring(0, maxCount - str.length);
        return str;
    }
}
const secp256k1 = {}
secp256k1.getKeyPair = function () {
    var pair = ec.genKeyPair()

    return {
        privKey: buf2hex(pair.getPrivate().toArrayLike(Buffer, "be", 32)),
        pubKey: buf2hex(pair.getPublic().encode(true, true))
    }
}
secp256k1.getPk = function (sk) {
    var key = ec.keyPair({
        priv: sk,
        privEnc: 'hex',
    });
    return buf2hex(key.getPublic().encode(true, true))
}
secp256k1.sign = function (msg, sk) {
    var key = ec.keyPair({
        priv: sk,
        privEnc: 'hex',
    });

    var sign = key.sign(msg)

    var r = buf2hex(sign.r.toArrayLike(Buffer, "be", 32))
    var s = buf2hex(sign.s.toArrayLike(Buffer, "be", 32))
    var result = r + s + "0" + s.recoveryParam
    return result
}
secp256k1.verify = function (msg, sig) {
    var pk = secp256k1.recoverPubKey(msg, sig)
    if (!Buffer.isBuffer(sig)) {
        sig = new Buffer(sig, "hex");
    }
    var sigObj = { r: sig.slice(0, 32), s: sig.slice(32, 64) }
    return ec.verify(msg, sigObj, pk)
}
secp256k1.recoverPubKey = function (msg, sig) {
    if (!Buffer.isBuffer(sig)) {
        try {
            sig = new Buffer(sig, "hex");
        } catch (e) {
            console.log(sig)
        }
    }
    if (!Buffer.isBuffer(msg)) {
        msg = new Buffer(msg, "hex");
    }

    var sigObj = { r: sig.slice(0, 32), s: sig.slice(32, 64) }
    if (sig[64] === 0 || sig[64] === 1) {
        var p = ec.recoverPubKey(msg, sigObj, sig[64])
    } else {
        var p = ec.recoverPubKey(msg, sigObj, 0)
    }
    return p.encode(true, true)
}

const numberWithCommas = function (x) {
    return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
}

function buf2hex(buffer) { // buffer is an ArrayBuffer
    // var r = Array.prototype.map.call(new Uint8Array(buffer), x => ('00' + x.toString(16)).slice(-2));
    // var r = new Array(buffer.length)
    // buffer.map(function (x,i) {r[i]=('00' + x.toString(16)).slice(-2)});
    var r = new Array(buffer.length)
    for (var i = 0; i < buffer.length; i++) {
        var x = buffer[i]
        r[i] = ('00' + x.toString(16)).slice(-2)
    }
    return r.join('')
}

function pkToPHash(pubKey) {
    if (!Buffer.isBuffer(pubKey)) {
        pubKey = new Buffer(pubKey, "hex")
    }

    var dh = DoubleHash(pubKey)

    var c = Checksum(pubKey)
    var r = new Buffer(31)
    r[0] = c
    for (var i = 1; i < r.length; i++) {
        r[i] = dh[i - 1]
    }

    return bs58.encode(r)
}

function skToPHash(privKey) {
    if (Buffer.isBuffer(privKey)) {
        privKey = buf2hex(privKey)
    }
    var pubKey = secp256k1.getPk(privKey)
    return pkToPHash(pubKey)
}

function verify(hash, sig, pubKey) {
    if (Buffer.isBuffer(hash)) {
        hash = buf2hex(hash)
    }
    if (Buffer.isBuffer(sig)) {
        sig = buf2hex(sig)
    }
    if (Buffer.isBuffer(pubKey)) {
        pubKey = buf2hex(pubKey)
    }
    // log("verify hash : " + hash
    //     + " sig : " + sig
    //     + " pubKey : " +  pubKey
    //     + " result : " + secp256k1.verify(hash, sig, pubKey))

    return secp256k1.verify(hash, sig, pubKey)
}

function ReadWithSize(buf, start, size, r) {
    var t = buf.slice(start, start + size)
    return { v: t, n: size }
}

function Read(buf, start, r) {
    var n = 0;
    r = ReadUint8(buf, start + n);
    var Len = r.v;
    n += r.n;

    var bs;
    if (Len < 254) {
        r = ReadWithSize(buf, start + n, Len);
        bs = r.v;
        n += r.n;
        return { r: bs, n: n }
    } else if (Len == 254) {
        r = ReadUint16(buf, start + n);
        Len = r.v;
        n += r.n;

        r = ReadWithSize(buf, start + n, Len);
        bs = r.v;
        n += r.n;
        return { r: bs, n: n }

    } else {
        r = ReadUint32(buf, start + n);
        Len = r.v;
        n += r.n;

        r = ReadWithSize(buf, start + n, Len);
        bs = r.v;
        n += r.n;
        return { r: bs, n: n }
    }
}

function ReadUint8(buf, start) {
    var v = buf.readUIntLE(start, 1)
    return { v: v, n: 1 }
}
function ReadUint16(buf, start) {
    var v = buf.readUIntLE(start, 2)
    return { v: v, n: 2 }
}
function ReadUint32(buf, start) {
    var v = buf.readUIntLE(start, 4)
    return { v: v, n: 4 }
}
function ReadUint64(buf, start) {
    var v = buf.readUIntLE(start, 8)
    return { v: v, n: 8 }
}
function ReadUint(buf, start, size) {
    var end = start + size
    var u8 = buf.slice(start, end); // original array
    var ubytes = u8.buffer.slice(start, end); // last four bytes as a new `ArrayBuffer`
    return ubytes
}

function WriteUint8(buf, start, value) {
    if (!Buffer.isBuffer(buf)) {
        throw "is not a Buffer"
    }
    buf.writeUIntLE(value, start, 1)
    return 1
}
function WriteUint16(buf, start, value) {
    if (!Buffer.isBuffer(buf)) {
        throw "is not a Buffer"
    }

    buf.writeUIntLE(value, start, 2)
    return 2
}
function WriteUint32(buf, start, value) {
    if (!Buffer.isBuffer(buf)) {
        throw "is not a Buffer"
    }

    buf.writeUIntLE(value, start, 4)
    return 4
}
function WriteUint64(buf, start, value) {
    if (!Buffer.isBuffer(buf)) {
        throw "is not a Buffer"
    }
    if (value > Number.MAX_SAFE_INTEGER) {
        throw "javascript cannot handle values over the " + Number.MAX_SAFE_INTEGER
    }

    buf.writeUIntLE(value, start, 8)
    return 8
}
function writeFixedBytes(buf, start, bs) {
    if (!Buffer.isBuffer(bs)) {
        throw "is not a Buffer"
    }
    for (var i = 0; i < bs.length; i++) {
        buf[i + start] = bs[i]
    }
    return bs.length
}
function WriteBytes(buf, start, bs) {
    if (typeof bs === "string") {
        bs = new Buffer(bs)
    }
    if (!Buffer.isBuffer(bs)) {
        throw "is not a Buffer"
    }
    var n = 0
    if (bs.length < 254) {
        if (typeof buf === "undefined") {
            start = 0
            buf = new Buffer(1 + bs.length)
        }

        n += WriteUint8(buf, start + n, bs.length)
        n += writeFixedBytes(buf, start + n, bs)
    } else if (bs.length < 65536) {
        if (typeof buf === "undefined") {
            start = 0
            buf = new Buffer(3 + bs.length)
        }

        n += WriteUint8(buf, start + n, 254)
        n += WriteUint16(buf, start + n, bs.length)
        n += writeFixedBytes(buf, start + n, bs)
    } else {
        if (typeof buf === "undefined") {
            start = 0
            buf = new Buffer(5 + bs.length)
        }

        n += WriteUint8(buf, start + n, 255)
        n += WriteUint32(buf, start + n, bs.length)
        n += writeFixedBytes(buf, start + n, bs)
    }
    return { v: buf, n: n }
}

function DoubleHash(v) {
    if (Buffer.isBuffer(v) !== true) {
        throw "param is not Buffer"
    }
    var msg = sha256(v);
    msg = sha256(new Buffer(msg, "hex"));
    msg = new Buffer(msg, "hex")

    return msg
}

function Checksum(buf) {
    var b8 = new Uint8Array(1)

    for (var i = 0; i < buf.length; i++) {
        b8[0] = b8[0] ^ buf[i]
    }
    return b8[0]
}