# TOTP Security Implementation

## Overview

GoConnect encrypts TOTP secrets at rest using **AES-256-GCM** encryption to protect against database leaks.

## Why Encrypt TOTP Secrets?

### Attack Scenario Without Encryption
1. Attacker gains database access (SQL injection, backup leak, etc.)
2. Attacker extracts plaintext TOTP secrets
3. Attacker can generate valid TOTP codes for any user
4. **2FA is completely bypassed**

### Protection With Encryption
1. Attacker gains database access
2. Attacker only sees encrypted secrets (useless without encryption key)
3. **2FA remains secure** even if database is compromised

## Encryption Implementation

### Algorithm: AES-256-GCM

**Why AES-256-GCM?**
- **AES-256**: Industry standard, unbreakable with current technology
- **GCM Mode**: Authenticated encryption (prevents tampering)
- **Nonce**: Each encryption uses unique random nonce
- **No IV reuse**: GCM automatically handles nonce generation

### Data Flow

#### **Enrollment (Storing Secret)**
```
TOTP Secret (plaintext) 
  → Generate random nonce
  → AES-256-GCM encryption with key
  → Base64 encode
  → Store in database
```

#### **Verification (Retrieving Secret)**
```
Encrypted secret from database
  → Base64 decode
  → Extract nonce
  → AES-256-GCM decryption with key
  → TOTP Secret (plaintext)
  → Generate TOTP code
```

## Key Management

### 1. **Key Generation**

Generate a secure 32-byte (256-bit) key:

```bash
# Windows
.\scripts\generate-encryption-key.bat

# Linux/Mac
chmod +x scripts/generate-encryption-key.sh
./scripts/generate-encryption-key.sh
```

### 2. **Key Storage**

**❌ NEVER:**
- Commit key to Git
- Store in code
- Share via email/chat
- Store in database

**✅ DO:**
- Store in environment variables
- Use secrets manager (AWS Secrets Manager, Azure Key Vault, HashiCorp Vault)
- Rotate keys periodically
- Backup encrypted key separately

### 3. **Key Rotation**

When rotating encryption keys:

```go
// 1. Generate new key
newKey := generateNewKey()

// 2. Decrypt with old key, re-encrypt with new key
oldEncryptor := crypto.NewEncryptor(oldKey)
newEncryptor := crypto.NewEncryptor(newKey)

for each user with TOTP {
    secret := oldEncryptor.DecryptTOTPSecret(user.TOTPSecret)
    newEncrypted := newEncryptor.EncryptTOTPSecret(secret)
    updateDatabase(user.ID, newEncrypted)
}

// 3. Update TOTP_ENCRYPTION_KEY environment variable
```

## Security Best Practices

### 1. **Defense in Depth**
Encryption is ONE layer. Also implement:
- Database access controls
- Network segmentation
- Regular security audits
- Least privilege access
- Encrypted backups

### 2. **Redis Security for TOTP**
```go
// Prevent replay attacks
redis.SetNX("totp:used:{userID}:{code}", "1", 90*time.Second)

// Rate limiting
redis.Incr("totp:attempts:{userID}")
redis.Expire("totp:attempts:{userID}", 5*time.Minute)

// Block after 5 failed attempts
if attempts > 5 {
    redis.Set("totp:blocked:{userID}", "1", 15*time.Minute)
}
```

### 3. **Audit Logging**
Log all TOTP operations:
- Secret generation
- Verification attempts (success/failure)
- Enable/disable events
- Encryption/decryption operations

### 4. **Backup Security**
- **Database backups**: Already encrypted (secrets are encrypted)
- **Encryption key backups**: Store separately from database
- **Never** backup key in same location as database

## Threat Model

| Threat | Mitigation |
|--------|-----------|
| **Database leak** | ✅ Secrets encrypted |
| **Backup theft** | ✅ Secrets encrypted |
| **SQL injection** | ✅ Secrets encrypted |
| **Memory dump** | ⚠️ Secrets briefly in memory during use |
| **Key compromise** | ❌ All secrets exposed (rotate key immediately) |
| **Insider threat** | ✅ Key stored separately from DB |

## Production Deployment

### Environment Variables
```bash
# Production .env (NEVER commit this file)
TOTP_ENCRYPTION_KEY=<32-character-random-key>
```

### Docker Secrets
```yaml
services:
  auth-service:
    secrets:
      - totp_encryption_key
    environment:
      TOTP_ENCRYPTION_KEY_FILE: /run/secrets/totp_encryption_key

secrets:
  totp_encryption_key:
    external: true
```

### Kubernetes Secrets
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: totp-encryption-key
type: Opaque
data:
  key: <base64-encoded-32-char-key>
```

```yaml
# In deployment
env:
  - name: TOTP_ENCRYPTION_KEY
    valueFrom:
      secretKeyRef:
        name: totp-encryption-key
        key: key
```

## Compliance

This implementation helps meet:
- **GDPR**: Data protection by design
- **PCI DSS**: Strong cryptography for sensitive data
- **SOC 2**: Encryption at rest
- **HIPAA**: Technical safeguards

## Testing

### Unit Tests
```go
func TestTOTPEncryption(t *testing.T) {
    key := "12345678901234567890123456789012" // 32 chars
    encryptor, _ := crypto.NewEncryptor(key)
    
    secret := "JBSWY3DPEHPK3PXP"
    encrypted, _ := encryptor.EncryptTOTPSecret(secret)
    decrypted, _ := encryptor.DecryptTOTPSecret(encrypted)
    
    assert.Equal(t, secret, decrypted)
    assert.NotEqual(t, secret, encrypted)
}
```

### Security Tests
- Verify different nonces for each encryption
- Verify tampered ciphertext is rejected
- Verify wrong key cannot decrypt
- Verify key length validation

## Monitoring

Monitor for:
- Failed decryption attempts (possible corruption/tampering)
- High TOTP verification failure rates
- Unusual TOTP enablement patterns
- Key access patterns

## FAQ

**Q: What if I lose the encryption key?**  
A: All encrypted TOTP secrets become unrecoverable. Users must re-enroll in 2FA.

**Q: Can I use a shorter key?**  
A: No. AES-256 requires exactly 32 bytes. Shorter keys use weaker AES-128.

**Q: Should I encrypt the entire database?**  
A: Database-level encryption (TDE) is complementary, not a replacement. Use both.

**Q: How often should I rotate keys?**  
A: Every 90-365 days depending on security requirements.

**Q: What if encryption adds latency?**  
A: AES-GCM is hardware-accelerated on modern CPUs. Impact is negligible (<1ms).

## References

- [RFC 6238 - TOTP](https://datatracker.ietf.org/doc/html/rfc6238)
- [NIST SP 800-38D - GCM Mode](https://csrc.nist.gov/publications/detail/sp/800-38d/final)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
