pragma solidity ^0.4.25;

contract CrossChainOracle {
    // 专门抛给 Chainlink 节点监听的跨链事件
    event CertVerificationRequested(bytes32 indexed reqId, string certHash);

    // 存储跨链回写的结果
    mapping(bytes32 => bool) public certResults;
    uint256 private nonce = 0;

    // 1. 业务端调用此方法发起查证
    function requestVerify(string _hash) public returns (bytes32) {
        nonce++;
        // 生成唯一的任务流水号
        bytes32 reqId = keccak256(abi.encodePacked(msg.sender, nonce));
        
        // 触发事件，唤醒 Chainlink
        emit CertVerificationRequested(reqId, _hash);
        return reqId;
    }

    // 2. Chainlink 节点拿到 Fabric 结果后，调用此方法回写上链
    function fulfillVerification(bytes32 _reqId, bool _isValid) public {
        certResults[_reqId] = _isValid;
    }
    
    // 查询结果的辅助方法
    function getResult(bytes32 _reqId) public view returns (bool) {
        return certResults[_reqId];
    }
}
