pragma solidity ^0.4.25;

contract CrossChainClient {
    // 事件全用明文，方便外部监听
    event CertVerificationRequested(string certHash);
    event CertVerified(string certHash, bool isValid);

    // 🛡️ 核心存储（内刚）：只存 bytes32 哈希，极致节省 Gas 费！
    mapping(bytes32 => bool) private verifyResults;

    // 发起验证（外柔）：接收明文 string
    function requestVerification(string memory _certHash) public {
        emit CertVerificationRequested(_certHash);
    }

    // 回写验证：接收明文 string，在链上内部转哈希！(彻底解决 console 报错)
    function fulfillVerification(string memory _certHash, bool _isValid) public {
        bytes32 certId = keccak256(abi.encodePacked(_certHash));
        verifyResults[certId] = _isValid;
        emit CertVerified(_certHash, _isValid);
    }

    // 提供一个友好的明文查询接口
    function getResult(string memory _certHash) public view returns (bool) {
        bytes32 certId = keccak256(abi.encodePacked(_certHash));
        return verifyResults[certId];
    }
}
