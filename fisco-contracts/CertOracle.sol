pragma solidity ^0.4.25;

contract CertOracle {
    mapping(string => bool) certStatus;
    mapping(string => bool) isVerified;
    
    event CertVerificationRequested(string ipfsHash);

    function requestCrossChainVerification(string _ipfsHash) public {
        if (!isVerified[_ipfsHash]) {
            emit CertVerificationRequested(_ipfsHash);
        }
    }

    function writeBackCertStatus(string _ipfsHash, bool _isValid) public {
        certStatus[_ipfsHash] = _isValid;
        isVerified[_ipfsHash] = true;
    }
    
    function getCertStatus(string _ipfsHash) public view returns (bool, bool) {
        return (isVerified[_ipfsHash], certStatus[_ipfsHash]);
    }
}
