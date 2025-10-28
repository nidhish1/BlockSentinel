from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from datetime import datetime
import logging
from typing import Optional

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="Transaction Risk Analyzer",
    description="AI-powered blockchain transaction risk assessment",
    version="1.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Request/Response models
class Transaction(BaseModel):
    hash: str
    from_addr: str
    to: str
    value: str
    gas: int
    gasPrice: str
    blockNum: int
    timestamp: int
    input: str

class RiskAnalysis(BaseModel):
    risk_score: float
    risk_level: str
    reasoning: str
    transaction_hash: str
    timestamp: str
    confidence: float

class HealthResponse(BaseModel):
    status: str
    version: str
    timestamp: str

# Ultra-simple risk analyzer (no external dependencies)
class LightweightRiskAnalyzer:
    def __init__(self):
        self.suspicious_patterns = {
            "mixers": [
                "0x94A1B5CdB22c43faab4AbEb5c74999895464Ddaf".lower(),
                "0x12D66f87A04A9E220743712cE6d9bB1B5616B8Fc".lower(),
            ],
            "known_scams": [
                "0x0000000000000000000000000000000000000000".lower(),
            ],
            "high_risk_dexes": [
                "0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D".lower(),
            ]
        }
    
    def wei_to_eth(self, wei_str: str) -> float:
        """Convert wei string to ETH"""
        try:
            wei = int(wei_str)
            return wei / 1000000000000000000  # 10^18
        except:
            return 0.0
    
    def analyze_transaction(self, tx: Transaction) -> RiskAnalysis:
        risk_factors = []
        score = 0.0
        
        # Convert values to lowercase for comparison
        from_addr_lower = tx.from_addr.lower()
        to_addr_lower = tx.to.lower()
        
        # Convert values
        value_eth = self.wei_to_eth(tx.value)
        
        # Parse gas price (handle both numeric and hex)
        gas_price_gwei = 0
        try:
            if tx.gasPrice.startswith('0x'):
                gas_price_gwei = int(tx.gasPrice, 16) / 1000000000  # 10^9
            else:
                gas_price_gwei = int(tx.gasPrice) / 1000000000
        except:
            gas_price_gwei = 0
        
        logger.info(f"Analyzing tx {tx.hash}: {value_eth:.4f} ETH, {gas_price_gwei:.0f} Gwei")
        
        # 1. Value-based risk analysis
        if value_eth > 10:  # Large transfer
            risk_factors.append(f"Large transfer: {value_eth:.2f} ETH")
            score += 0.4
        elif value_eth > 1:  # Moderate transfer
            risk_factors.append(f"Moderate transfer: {value_eth:.2f} ETH")
            score += 0.2
        elif value_eth < 0.0001:  # Very small transfer
            risk_factors.append(f"Very small transfer: {value_eth:.6f} ETH")
            score += 0.1
        
        # 2. Gas price analysis
        if gas_price_gwei > 100:  # Very high gas price
            risk_factors.append(f"High gas price: {gas_price_gwei:.0f} Gwei")
            score += 0.3
        elif gas_price_gwei < 5:  # Suspiciously low
            risk_factors.append(f"Unusually low gas: {gas_price_gwei:.0f} Gwei")
            score += 0.15
        
        # 3. Gas limit analysis
        if tx.gas > 100000:  # High gas limit
            risk_factors.append(f"High gas limit: {tx.gas}")
            score += 0.1
        
        # 4. Known address patterns (high risk)
        if to_addr_lower in self.suspicious_patterns["mixers"]:
            risk_factors.append("Transaction to known mixer (Tornado Cash)")
            score += 0.8
        elif to_addr_lower in self.suspicious_patterns["known_scams"]:
            risk_factors.append("Transaction to burn address")
            score += 0.3
        
        if from_addr_lower in self.suspicious_patterns["mixers"]:
            risk_factors.append("Transaction from known mixer")
            score += 0.7
        
        # 5. Contract interaction complexity
        if tx.input and tx.input != "0x":
            input_length = len(tx.input)
            risk_factors.append("Smart contract interaction")
            score += 0.1
            
            if input_length > 500:
                risk_factors.append("Complex contract call")
                score += 0.15
            elif input_length > 1000:
                risk_factors.append("Very complex contract call")
                score += 0.25
        
        # 6. Behavioral patterns
        if value_eth > 1 and gas_price_gwei > 100:
            risk_factors.append("High-value tx with premium gas")
            score += 0.2
        
        # 7. Zero-value contract interactions
        if value_eth == 0 and tx.input != "0x":
            risk_factors.append("Contract interaction with 0 ETH value")
            score += 0.1
        
        # Determine risk level
        if score >= 0.7:
            level = "critical"
            confidence = min(score + 0.2, 0.95)
        elif score >= 0.5:
            level = "high" 
            confidence = min(score + 0.15, 0.9)
        elif score >= 0.3:
            level = "medium"
            confidence = min(score + 0.1, 0.8)
        elif score >= 0.1:
            level = "low"
            confidence = min(score + 0.05, 0.7)
        else:
            level = "very low"
            confidence = 0.6
        
        reasoning = " | ".join(risk_factors) if risk_factors else "Normal transaction patterns"
        
        logger.info(f"Risk analysis: {level} (score: {score:.2f}) - {reasoning}")
        
        return RiskAnalysis(
            risk_score=min(score, 1.0),
            risk_level=level,
            reasoning=reasoning,
            transaction_hash=tx.hash,
            timestamp=datetime.utcnow().isoformat(),
            confidence=round(confidence, 2)
        )

# Initialize analyzer
analyzer = LightweightRiskAnalyzer()

# API Routes
@app.post("/analyze", response_model=RiskAnalysis)
async def analyze_transaction(tx: Transaction):
    """
    Analyze transaction for risk factors
    """
    try:
        logger.info(f"Analyzing transaction: {tx.hash}")
        analysis = analyzer.analyze_transaction(tx)
        logger.info(f"Risk analysis complete: {tx.hash} -> {analysis.risk_level} ({analysis.risk_score})")
        return analysis
    except Exception as e:
        logger.error(f"Error analyzing transaction {tx.hash}: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Analysis failed: {str(e)}")

@app.get("/health", response_model=HealthResponse)
async def health_check():
    """
    Health check endpoint
    """
    return HealthResponse(
        status="healthy",
        version="1.0.0",
        timestamp=datetime.utcnow().isoformat()
    )

@app.get("/")
async def root():
    return {
        "message": "Transaction Risk Analyzer API",
        "version": "1.0.0",
        "endpoints": {
            "analyze": "POST /analyze - Analyze transaction risk",
            "health": "GET /health - Service health check"
        }
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000, log_level="info")