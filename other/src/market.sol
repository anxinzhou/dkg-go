pragma solidity ^0.5.7;

contract Market {

    address [] public committees;
    uint256 constant public r = 2;
    uint256 constant public timestamp = 1555295166636956000;
    bytes32 constant difficulty = 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff;
    uint256 constant reward =2;
    bytes32 [] public knowledges;
    address [] public servers;
    bytes32 public buyer;
    bytes32 [] public decryptionShare;

    mapping(address=>uint256) balance;

    constructor () public {
        balance[msg.sender] = 100000000;
    }

    function balanceOf(address user)  external view returns (uint256) {
        return balance[user];
    }

    function transfer(address from, address to, uint256 value) public {
        require(balance[from]>=value);
        balance[to]+=value;
        balance[from]-=value;
    }

    function registerCommittee(uint256 nonce) external {
        address user = msg.sender;
        bytes memory v = abi.encodePacked(bytes32(bytes20(user)) | bytes32(r) | difficulty | bytes32(nonce));
        bytes32 hashV = keccak256(v);
        require(hashV<difficulty);
        committees.push(msg.sender);
    }

    function submitKnowledge(bytes32 cipher) external {
        require(knowledges.length<2);
        knowledges.push(cipher);
        servers.push(msg.sender);
    }

    function buyKnowledge(bytes32 publicKey) external {
        transfer(msg.sender,servers[0],reward);
        transfer(msg.sender,servers[1],reward);
        buyer = publicKey;
    }

    function submitDecryptionShare(bytes32 share) external {
        decryptionShare.push(share);
        transfer(msg.sender,servers[0],reward);
    }
}