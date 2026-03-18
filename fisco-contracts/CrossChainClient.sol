pragma solidity ^0.4.25;

contract CrossChainClient {
    // 抛出跨链请求事件，预言机节点会监听这个事件
    event CertVerificationRequested(bytes32 indexed reqId, string certHash);
    // 跨链结果回调事件
    event CertVerified(bytes32 indexed reqId, bool isValid);

    // 存储跨链验证结果
    mapping(bytes32 => bool) public verifyResults;

    // 业务端调用此方法发起验证
    function requestVerification(string memory _certHash) public returns (bytes32) {
        bytes32 reqId = keccak256(abi.encodePacked(msg.sender, block.timestamp, _certHash));
        emit CertVerificationRequested(reqId, _certHash);
        return reqId;
    }

    // 预言机获取到 Fabric 结果后，调用此方法将数据写回 FISCO
    function fulfillVerification(bytes32 _reqId, bool _isValid) public {
        verifyResults[_reqId] = _isValid;
        emit CertVerified(_reqId, _isValid);
    }
}
