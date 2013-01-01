package TLibCommon

import (
	"math"
	//"fmt"
)

// ====================================================================================================================
// Constants
// ====================================================================================================================
const SAO_MAX_DEPTH = 4
const SAO_BO_BITS = 5
const LUMA_GROUP_NUM = (1 << SAO_BO_BITS)
const MAX_NUM_SAO_OFFSETS = 4
const MAX_NUM_SAO_CLASS = 33


var m_aiNumCulPartsLevel=[5]int{
  1,   //level 0
  5,   //level 1
  21,  //level 2
  85,  //level 3
  341, //level 4
};

var m_auiEoTable=[9]uint{
  1, //0    
  2, //1   
  0, //2
  3, //3
  4, //4
  0, //5  
  0, //6  
  0, //7 
  0,
};

var m_iNumClass=[MAX_NUM_SAO_TYPE]int{
  SAO_EO_LEN,
  SAO_EO_LEN,
  SAO_EO_LEN,
  SAO_EO_LEN,
  SAO_BO_LEN,
};

const m_uiMaxDepth = SAO_MAX_DEPTH;

// ====================================================================================================================
// Class definition
// ====================================================================================================================

/// Sample Adaptive Offset class
type TComSampleAdaptiveOffset struct {
    //protected:
    m_pcPic *TComPic

    //m_uiMaxDepth         uint
    //m_aiNumCulPartsLevel [5]uint
    //m_auiEoTable         [9]uint
    m_iOffsetBo          []int
    m_iChromaOffsetBo    []int
    m_iOffsetEo          [LUMA_GROUP_NUM]int

    m_iPicWidth           int
    m_iPicHeight          int
    m_uiMaxSplitLevel     uint
    m_uiMaxCUWidth        uint
    m_uiMaxCUHeight       uint
    m_iNumCuInWidth       int
    m_iNumCuInHeight      int
    m_iNumTotalParts      int
    //m_iNumClass           [MAX_NUM_SAO_TYPE]int
    m_eSliceType          SliceType
    m_iPicNalReferenceIdc int

    m_uiSaoBitIncreaseY uint
    m_uiSaoBitIncreaseC uint //for chroma
    m_uiQP              uint

    m_pClipTable           []Pel
    m_pClipTableBase       []Pel
    m_lumaTableBo          []Pel
    m_pChromaClipTable     []Pel
    m_pChromaClipTableBase []Pel
    m_chromaTableBo        []Pel
    m_iUpBuff1             []int
    //m_iUpBuff2             []int
    m_iUpBufft             []int
    m_ipSwap               []int
    m_bUseNIF              bool        //!< true for performing non-cross slice boundary ALF
    m_uiNumSlicesInPic     uint        //!< number of slices in picture
    m_iSGDepth             int         //!< slice granularity depth
    m_pcYuvTmp             *TComPicYuv //!< temporary picture buffer pointer when non-across slice/tile boundary SAO is enabled

    m_pTmpU1                  []Pel
    m_pTmpU2                  []Pel
    m_pTmpL1                  []Pel
    m_pTmpL2                  []Pel
    m_iLcuPartIdx             []int
    m_maxNumOffsetsPerPic     int
    m_saoLcuBoundary          bool
    m_saoLcuBasedOptimization bool
}


//public:
func NewTComSampleAdaptiveOffset() *TComSampleAdaptiveOffset{
	return &TComSampleAdaptiveOffset{}
}

func (this *TComSampleAdaptiveOffset) Create(uiSourceWidth, uiSourceHeight, uiMaxCUWidth, uiMaxCUHeight uint) {
  this.m_iPicWidth  = int(uiSourceWidth);
  this.m_iPicHeight = int(uiSourceHeight);

  this.m_uiMaxCUWidth  = uiMaxCUWidth;
  this.m_uiMaxCUHeight = uiMaxCUHeight;

  this.m_iNumCuInWidth  = this.m_iPicWidth / int(this.m_uiMaxCUWidth);
  this.m_iNumCuInWidth += int(B2U(( this.m_iPicWidth % int(this.m_uiMaxCUWidth) )!=0));// ? 1 : 0;

  this.m_iNumCuInHeight  = this.m_iPicHeight / int(this.m_uiMaxCUHeight);
  this.m_iNumCuInHeight += int(B2U(( this.m_iPicHeight % int(this.m_uiMaxCUHeight) )!=0));// ? 1 : 0;

  iMaxSplitLevelHeight := int(math.Log2(float64(this.m_iNumCuInHeight)));//(Int)(logf((Float)this.m_iNumCuInHeight)/logf(2.0));
  iMaxSplitLevelWidth  := int(math.Log2(float64(this.m_iNumCuInWidth )));//(Int)(logf((Float)this.m_iNumCuInWidth )/logf(2.0));

  if iMaxSplitLevelHeight < iMaxSplitLevelWidth {
  	this.m_uiMaxSplitLevel = uint(iMaxSplitLevelHeight);
  }else{
  	this.m_uiMaxSplitLevel = uint(iMaxSplitLevelWidth);
  }
  if this.m_uiMaxSplitLevel< m_uiMaxDepth {
  	this.m_uiMaxSplitLevel = this.m_uiMaxSplitLevel;
  }else{
  	this.m_uiMaxSplitLevel = m_uiMaxDepth;
  }
  /* various structures are overloaded to store per component data.
   * this.m_iNumTotalParts must allow for sufficient storage in any allocated arrays */
  this.m_iNumTotalParts  = MAX(3,m_aiNumCulPartsLevel[this.m_uiMaxSplitLevel]).(int);

  uiPixelRangeY   := 1 << uint(G_bitDepthY);
  uiBoRangeShiftY := uint(G_bitDepthY - SAO_BO_BITS);

  this.m_lumaTableBo = make([]Pel, uiPixelRangeY);
  for k2:=0; k2<uiPixelRangeY; k2++ {
    this.m_lumaTableBo[k2] = 1 + Pel(k2>>uiBoRangeShiftY);
  }

  uiPixelRangeC   := 1 << uint(G_bitDepthC);
  uiBoRangeShiftC := uint(G_bitDepthC - SAO_BO_BITS);

  this.m_chromaTableBo = make([]Pel, uiPixelRangeC);
  for k2:=0; k2<uiPixelRangeC; k2++ {
    this.m_chromaTableBo[k2] = 1 + Pel(k2>>uiBoRangeShiftC);
  }

  this.m_iUpBuff1 = make([]int, this.m_iPicWidth+2);
  //this.m_iUpBuff2 = make([]int, this.m_iPicWidth+2);
  this.m_iUpBufft = make([]int, this.m_iPicWidth+2);

  //fmt.Printf("Potential bug\n")
  //this.m_iUpBuff1++;
  //this.m_iUpBuff2++;
  //this.m_iUpBufft++;
  
  var i int;

  uiMaxY  := int(1 << uint(G_bitDepthY)) - 1;;
  uiMinY  := int(0);

  iCRangeExt := int(uiMaxY)>>1;

  this.m_pClipTableBase = make([]Pel, uiMaxY+2*iCRangeExt);
  this.m_iOffsetBo      = make([]int, uiMaxY+2*iCRangeExt);

  for i=0;i<(uiMinY+iCRangeExt);i++ {
    this.m_pClipTableBase[i] = Pel(uiMinY);
  }

  for i=uiMinY+iCRangeExt;i<(uiMaxY+iCRangeExt);i++ {
    this.m_pClipTableBase[i] = Pel(i-iCRangeExt);
  }

  for i=uiMaxY+iCRangeExt;i<(uiMaxY+2*iCRangeExt);i++ {
    this.m_pClipTableBase[i] = Pel(uiMaxY);
  }

  this.m_pClipTable = this.m_pClipTableBase[iCRangeExt:];

  uiMaxC  := int(1 << uint(G_bitDepthC)) - 1;
  uiMinC  := int(0);

  iCRangeExtC := int(uiMaxC)>>1;

  this.m_pChromaClipTableBase = make([]Pel, uiMaxC+2*iCRangeExtC);
  this.m_iChromaOffsetBo      = make([]int, uiMaxC+2*iCRangeExtC);

  for i=0;i<(uiMinC+iCRangeExtC);i++ {
    this.m_pChromaClipTableBase[i] = Pel(uiMinC);
  }

  for i=uiMinC+iCRangeExtC;i<(uiMaxC+iCRangeExtC);i++ {
    this.m_pChromaClipTableBase[i] = Pel(i-iCRangeExtC);
  }

  for i=uiMaxC+iCRangeExtC;i<(uiMaxC+2*iCRangeExtC);i++ {
    this.m_pChromaClipTableBase[i] = Pel(uiMaxC);
  }

  this.m_pChromaClipTable = this.m_pChromaClipTableBase[iCRangeExtC:];

  this.m_iLcuPartIdx = make([]int, this.m_iNumCuInHeight*this.m_iNumCuInWidth);
  this.m_pTmpL1 = make([]Pel, this.m_uiMaxCUHeight+1);
  this.m_pTmpL2 = make([]Pel, this.m_uiMaxCUHeight+1);
  this.m_pTmpU1 = make([]Pel, this.m_iPicWidth);
  this.m_pTmpU2 = make([]Pel, this.m_iPicWidth);
}
func (this *TComSampleAdaptiveOffset) Destroy() {
  /*if (this.m_pClipTableBase)
  {
    delete [] this.m_pClipTableBase; this.m_pClipTableBase = NULL;
  }
  if (this.m_iOffsetBo)
  {
    delete [] this.m_iOffsetBo; this.m_iOffsetBo = NULL;
  }
  if (this.m_lumaTableBo)
  {
    delete[] this.m_lumaTableBo; this.m_lumaTableBo = NULL;
  }

  if (this.m_pChromaClipTableBase)
  {
    delete [] this.m_pChromaClipTableBase; this.m_pChromaClipTableBase = NULL;
  }
  if (this.m_iChromaOffsetBo)
  {
    delete [] this.m_iChromaOffsetBo; this.m_iChromaOffsetBo = NULL;
  }
  if (this.m_chromaTableBo)
  {
    delete[] this.m_chromaTableBo; this.m_chromaTableBo = NULL;
  }

  if (this.m_iUpBuff1)
  {
    this.m_iUpBuff1--;
    delete [] this.m_iUpBuff1; this.m_iUpBuff1 = NULL;
  }
  if (this.m_iUpBuff2)
  {
    this.m_iUpBuff2--;
    delete [] this.m_iUpBuff2; this.m_iUpBuff2 = NULL;
  }
  if (this.m_iUpBufft)
  {
    this.m_iUpBufft--;
    delete [] this.m_iUpBufft; this.m_iUpBufft = NULL;
  }
  if (this.m_pTmpL1)
  {
    delete [] this.m_pTmpL1; this.m_pTmpL1 = NULL;
  }
  if (this.m_pTmpL2)
  {
    delete [] this.m_pTmpL2; this.m_pTmpL2 = NULL;
  }
  if (this.m_pTmpU1)
  {
    delete [] this.m_pTmpU1; this.m_pTmpU1 = NULL;
  }
  if (this.m_pTmpU2)
  {
    delete [] this.m_pTmpU2; this.m_pTmpU2 = NULL;
  }
  if(this.m_iLcuPartIdx)
  {
    delete []this.m_iLcuPartIdx; this.m_iLcuPartIdx = NULL;
  }*/
}


func (this *TComSampleAdaptiveOffset) ConvertLevelRowCol2Idx( level,  row,  col int) int{
  var idx int;
  if level == 0 {
    idx = 0;
  }else if level == 1 {
    idx = 1 + row*2 + col;
  }else if level == 2 {
    idx = 5 + row*4 + col;
  }else if level == 3 {
    idx = 21 + row*8 + col;
  }else{ // (level == 4)
    idx = 85 + row*16 + col;
  }
  
  return idx;
}

func (this *TComSampleAdaptiveOffset) InitSAOParam   (pcSaoParam *SAOParam,  iPartLevel,  iPartRow,  iPartCol,  iParentPartIdx,  StartCUX,  EndCUX,  StartCUY,  EndCUY,  iYCbCr int){
  var j int;
  iPartIdx := this.ConvertLevelRowCol2Idx(iPartLevel, iPartRow, iPartCol);

  var pSaoPart *SAOQTPart;

  pSaoPart = &(pcSaoParam.SaoPart[iYCbCr][iPartIdx]);

  pSaoPart.PartIdx   = iPartIdx;
  pSaoPart.PartLevel = iPartLevel;
  pSaoPart.PartRow   = iPartRow;
  pSaoPart.PartCol   = iPartCol;

  pSaoPart.StartCUX  = StartCUX;
  pSaoPart.EndCUX    = EndCUX;
  pSaoPart.StartCUY  = StartCUY;
  pSaoPart.EndCUY    = EndCUY;

  pSaoPart.UpPartIdx = iParentPartIdx;
  pSaoPart.iBestType   = -1;
  pSaoPart.iLength     =  0;

  pSaoPart.subTypeIdx = 0;

  for j=0;j<MAX_NUM_SAO_OFFSETS;j++ {
    pSaoPart.iOffset[j] = 0;
  }

  if pSaoPart.PartLevel != int(this.m_uiMaxSplitLevel) {
    DownLevel    := (iPartLevel+1 );
    DownRowStart := (iPartRow << 1);
    DownColStart := (iPartCol << 1);

    var iDownRowIdx, iDownColIdx, NumCUWidth,  NumCUHeight, NumCULeft, NumCUTop int;

    var DownStartCUX, DownStartCUY, DownEndCUX, DownEndCUY int;

    NumCUWidth  = EndCUX - StartCUX +1;
    NumCUHeight = EndCUY - StartCUY +1;
    NumCULeft   = (NumCUWidth  >> 1);
    NumCUTop    = (NumCUHeight >> 1);

    DownStartCUX= StartCUX;
    DownEndCUX  = DownStartCUX + NumCULeft - 1;
    DownStartCUY= StartCUY;
    DownEndCUY  = DownStartCUY + NumCUTop  - 1;
    iDownRowIdx = DownRowStart + 0;
    iDownColIdx = DownColStart + 0;

    pSaoPart.DownPartsIdx[0]= this.ConvertLevelRowCol2Idx(DownLevel, iDownRowIdx, iDownColIdx);

    this.InitSAOParam(pcSaoParam, DownLevel, iDownRowIdx, iDownColIdx, iPartIdx, DownStartCUX, DownEndCUX, DownStartCUY, DownEndCUY, iYCbCr);

    DownStartCUX = StartCUX + NumCULeft;
    DownEndCUX   = EndCUX;
    DownStartCUY = StartCUY;
    DownEndCUY   = DownStartCUY + NumCUTop -1;
    iDownRowIdx  = DownRowStart + 0;
    iDownColIdx  = DownColStart + 1;

    pSaoPart.DownPartsIdx[1] = this.ConvertLevelRowCol2Idx(DownLevel, iDownRowIdx, iDownColIdx);

    this.InitSAOParam(pcSaoParam, DownLevel, iDownRowIdx, iDownColIdx, iPartIdx,  DownStartCUX, DownEndCUX, DownStartCUY, DownEndCUY, iYCbCr);

    DownStartCUX = StartCUX;
    DownEndCUX   = DownStartCUX + NumCULeft -1;
    DownStartCUY = StartCUY + NumCUTop;
    DownEndCUY   = EndCUY;
    iDownRowIdx  = DownRowStart + 1;
    iDownColIdx  = DownColStart + 0;

    pSaoPart.DownPartsIdx[2] = this.ConvertLevelRowCol2Idx(DownLevel, iDownRowIdx, iDownColIdx);

    this.InitSAOParam(pcSaoParam, DownLevel, iDownRowIdx, iDownColIdx, iPartIdx, DownStartCUX, DownEndCUX, DownStartCUY, DownEndCUY, iYCbCr);

    DownStartCUX = StartCUX+ NumCULeft;
    DownEndCUX   = EndCUX;
    DownStartCUY = StartCUY + NumCUTop;
    DownEndCUY   = EndCUY;
    iDownRowIdx  = DownRowStart + 1;
    iDownColIdx  = DownColStart + 1;

    pSaoPart.DownPartsIdx[3] = this.ConvertLevelRowCol2Idx(DownLevel, iDownRowIdx, iDownColIdx);

    this.InitSAOParam(pcSaoParam, DownLevel, iDownRowIdx, iDownColIdx, iPartIdx,DownStartCUX, DownEndCUX, DownStartCUY, DownEndCUY, iYCbCr);
  }else{
    pSaoPart.DownPartsIdx[0]= -1;
    pSaoPart.DownPartsIdx[1]= -1;
    pSaoPart.DownPartsIdx[2]= -1;
    pSaoPart.DownPartsIdx[3]= -1; 
  }
}
func (this *TComSampleAdaptiveOffset) AllocSaoParam  (pcSaoParam *SAOParam){
  pcSaoParam.MaxSplitLevel = int(this.m_uiMaxSplitLevel);
  pcSaoParam.SaoPart[0] = make([]SAOQTPart, m_aiNumCulPartsLevel[pcSaoParam.MaxSplitLevel] );
  this.InitSAOParam(pcSaoParam, 0, 0, 0, -1, 0, this.m_iNumCuInWidth-1,  0, this.m_iNumCuInHeight-1,0);
  pcSaoParam.SaoPart[1] = make([]SAOQTPart, m_aiNumCulPartsLevel[pcSaoParam.MaxSplitLevel] );
  pcSaoParam.SaoPart[2] = make([]SAOQTPart, m_aiNumCulPartsLevel[pcSaoParam.MaxSplitLevel] );
  this.InitSAOParam(pcSaoParam, 0, 0, 0, -1, 0, this.m_iNumCuInWidth-1,  0, this.m_iNumCuInHeight-1,1);
  this.InitSAOParam(pcSaoParam, 0, 0, 0, -1, 0, this.m_iNumCuInWidth-1,  0, this.m_iNumCuInHeight-1,2);
  pcSaoParam.NumCuInWidth  = this.m_iNumCuInWidth;
  pcSaoParam.NumCuInHeight = this.m_iNumCuInHeight;
  pcSaoParam.SaoLcuParam[0] = make([]SaoLcuParam, this.m_iNumCuInHeight*this.m_iNumCuInWidth);
  pcSaoParam.SaoLcuParam[1] = make([]SaoLcuParam, this.m_iNumCuInHeight*this.m_iNumCuInWidth);
  pcSaoParam.SaoLcuParam[2] = make([]SaoLcuParam, this.m_iNumCuInHeight*this.m_iNumCuInWidth);
}
func (this *TComSampleAdaptiveOffset) ResetSAOParam  (pcSaoParam *SAOParam){
  iNumComponet := 3;
  for c:=0; c<iNumComponet; c++ {
	if c<2{
		pcSaoParam.SaoFlag[c] = false;
	}
    for i:=0; i< m_aiNumCulPartsLevel[this.m_uiMaxSplitLevel]; i++ {
      pcSaoParam.SaoPart[c][i].iBestType     = -1;
      pcSaoParam.SaoPart[c][i].iLength       =  0;
      pcSaoParam.SaoPart[c][i].bSplit        = false; 
      pcSaoParam.SaoPart[c][i].bProcessed    = false;
      pcSaoParam.SaoPart[c][i].dMinCost      = MAX_DOUBLE;
      pcSaoParam.SaoPart[c][i].iMinDist      = MAX_INT;
      pcSaoParam.SaoPart[c][i].iMinRate      = MAX_INT;
      pcSaoParam.SaoPart[c][i].subTypeIdx    = 0;
      for j:=0;j<MAX_NUM_SAO_OFFSETS;j++ {
        pcSaoParam.SaoPart[c][i].iOffset[j] = 0;
        pcSaoParam.SaoPart[c][i].iOffset[j] = 0;
        pcSaoParam.SaoPart[c][i].iOffset[j] = 0;
      }
    }
    pcSaoParam.OneUnitFlag[0]   = false;
    pcSaoParam.OneUnitFlag[1]   = false;
    pcSaoParam.OneUnitFlag[2]   = false;
    this.ResetLcuPart(pcSaoParam.SaoLcuParam[0]);
    this.ResetLcuPart(pcSaoParam.SaoLcuParam[1]);
    this.ResetLcuPart(pcSaoParam.SaoLcuParam[2]);
  }
}
func (this *TComSampleAdaptiveOffset) FreeSaoParam   (pcSaoParam *SAOParam){
  /*delete [] pcSaoParam.psSaoPart[0];
  delete [] pcSaoParam.psSaoPart[1];
  delete [] pcSaoParam.psSaoPart[2];
  pcSaoParam.psSaoPart[0] = 0;
  pcSaoParam.psSaoPart[1] = 0;
  pcSaoParam.psSaoPart[2] = 0;
  if( pcSaoParam.saoLcuParam[0]) 
  {
    delete [] pcSaoParam.saoLcuParam[0]; pcSaoParam.saoLcuParam[0] = NULL;
  }
  if( pcSaoParam.saoLcuParam[1]) 
  {
    delete [] pcSaoParam.saoLcuParam[1]; pcSaoParam.saoLcuParam[1] = NULL;
  }
  if( pcSaoParam.saoLcuParam[2]) 
  {
    delete [] pcSaoParam.saoLcuParam[2]; pcSaoParam.saoLcuParam[2] = NULL;
  }*/
}

func (this *TComSampleAdaptiveOffset) SAOProcess(pcSaoParam *SAOParam){
  if pcSaoParam.SaoFlag[0] || pcSaoParam.SaoFlag[1] {
    this.m_uiSaoBitIncreaseY = uint(MAX(int(G_bitDepthY - 10), int(0)).(int));
    this.m_uiSaoBitIncreaseC = uint(MAX(int(G_bitDepthC - 10), int(0)).(int));

    if this.m_bUseNIF {
      this.m_pcPic.GetPicYuvRec().CopyToPic(this.m_pcYuvTmp);
    }
    if this.m_saoLcuBasedOptimization {
      pcSaoParam.OneUnitFlag[0] = false;  
      pcSaoParam.OneUnitFlag[1] = false;  
      pcSaoParam.OneUnitFlag[2] = false;  
    }
    iY  := 0;
    if pcSaoParam.SaoFlag[0] {
      this.ProcessSaoUnitAll( pcSaoParam.SaoLcuParam[iY], pcSaoParam.OneUnitFlag[iY], iY);
    }
    if pcSaoParam.SaoFlag[1] {
       this.ProcessSaoUnitAll( pcSaoParam.SaoLcuParam[1], pcSaoParam.OneUnitFlag[1], 1);//Cb
       this.ProcessSaoUnitAll( pcSaoParam.SaoLcuParam[2], pcSaoParam.OneUnitFlag[2], 2);//Cr
    }
    this.m_pcPic = nil;
  }
}
func (this *TComSampleAdaptiveOffset) ProcessSaoCu( iAddr,  iSaoType,  iYCbCr int){
  if !this.m_bUseNIF {
    this.ProcessSaoCuOrg( iAddr, iSaoType, iYCbCr);
  }else{  
    isChroma := B2U(iYCbCr != 0);
    var  stride  int;
    if iYCbCr != 0 {
    	stride = this.m_pcPic.GetCStride();
    }else{
    	stride = this.m_pcPic.GetStride();
    }
    pPicRest := this.GetPicYuvAddr(this.m_pcPic.GetPicYuvRec(), iYCbCr, 0);
    pPicDec  := this.GetPicYuvAddr(this.m_pcYuvTmp, iYCbCr, 0);

    //variables
    var xPos, yPos, width, height uint;
    var pbBorderAvail []bool;
    var posOffset uint;

	vFilterBlocksList := this.m_pcPic.GetCU(uint(iAddr)).GetNDBFilterBlocks();//std::vector<NDBFBlockInfo>& 
    //for i:=0; i< vFilterBlocks.size(); i++ {
    for e:=vFilterBlocksList.Front(); e!=nil; e=e.Next() {
      vFilterBlocks:=e.Value.(NDBFBlockInfo)
      xPos          = vFilterBlocks.posX   >> isChroma;
      yPos          = vFilterBlocks.posY   >> isChroma;
      width         = vFilterBlocks.width  >> isChroma;
      height        = vFilterBlocks.height >> isChroma;
      pbBorderAvail = vFilterBlocks.isBorderAvailable[:];

      posOffset = yPos* uint(stride) + xPos;

      this.ProcessSaoBlock(pPicDec[posOffset:], pPicRest[posOffset:], stride, iSaoType, int(width), int(height), pbBorderAvail, iYCbCr);
    }
  }
}
func (this *TComSampleAdaptiveOffset) GetPicYuvAddr(pcPicYuv *TComPicYuv,  iYCbCr, iAddr int) []Pel{
  switch iYCbCr {
  case 0:
    return pcPicYuv.GetLumaAddr1(iAddr);
    //break;
  case 1:
    return pcPicYuv.GetCbAddr1(iAddr);
    //break;
  case 2:
    return pcPicYuv.GetCrAddr1(iAddr);
    //break;
  default:
    return nil;
    //break;
  }
  return nil
}

func (this *TComSampleAdaptiveOffset) xSign(x int) int{
  return ((x >> 31) | (int( uint(-x) >> 31)));
}

func (this *TComSampleAdaptiveOffset) ProcessSaoCuOrg( iAddr,  iSaoType,  iYCbCr int){  //!< LCU-basd SAO process without slice granularity
  var x,y int;
  pTmpCu := this.m_pcPic.GetCU(uint(iAddr));
  var pRec []Pel;
  var  iStride int;
  iLcuWidth  := int(this.m_uiMaxCUWidth);
  iLcuHeight := int(this.m_uiMaxCUHeight);
  uiLPelX    := pTmpCu.GetCUPelX();
  uiTPelY    := pTmpCu.GetCUPelY();
  var uiRPelX, uiBPelY uint;
  var iSignLeft,iSignRight,iSignDown,iSignDown1,iSignDown2 int;
  var uiEdgeType uint;
  var iPicWidthTmp,iPicHeightTmp,iStartX,iStartY,iEndX,iEndY int;
  iIsChroma := B2U(iYCbCr!=0);
  var iShift,iCuHeightTmp int;
  var pTmpLSwap,pTmpL,pTmpU,pClipTbl []Pel;
  var pOffsetBo []int;

  iPicWidthTmp  = this.m_iPicWidth  >> iIsChroma;
  iPicHeightTmp = this.m_iPicHeight >> iIsChroma;
  iLcuWidth     = iLcuWidth    >> iIsChroma;
  iLcuHeight    = iLcuHeight   >> iIsChroma;
  uiLPelX       = uiLPelX      >> iIsChroma;
  uiTPelY       = uiTPelY      >> iIsChroma;
  uiRPelX       = uiLPelX + uint(iLcuWidth)  ;
  uiBPelY       = uiTPelY + uint(iLcuHeight) ;
  if uiRPelX > uint(iPicWidthTmp) {
  	uiRPelX       =   uint(iPicWidthTmp);
  }else{
  	uiRPelX       =   uiRPelX;
  }
  if uiBPelY > uint(iPicHeightTmp) {
  	uiBPelY       =  uint(iPicHeightTmp);
  }else{
  	uiBPelY       =  uiBPelY;
  }
  iLcuWidth     = int(uiRPelX - uiLPelX);
  iLcuHeight    = int(uiBPelY - uiTPelY);

  if pTmpCu.GetPic()==nil {
    return;
  }
  if iYCbCr == 0 {
    pRec       = this.m_pcPic.GetPicYuvRec().GetLumaAddr1(iAddr);
    iStride    = this.m_pcPic.GetStride();
  }else if iYCbCr == 1 {
    pRec       = this.m_pcPic.GetPicYuvRec().GetCbAddr1(iAddr);
    iStride    = this.m_pcPic.GetCStride();
  }else{
    pRec       = this.m_pcPic.GetPicYuvRec().GetCrAddr1(iAddr);
    iStride    = this.m_pcPic.GetCStride();
  }

//   if (iSaoType!=SAO_BO_0 || iSaoType!=SAO_BO_1)
  {
    iCuHeightTmp = int(this.m_uiMaxCUHeight >> iIsChroma);
    iShift = int(this.m_uiMaxCUWidth>> iIsChroma)-1;
    for i:=0;i<iCuHeightTmp+1;i++ {
      this.m_pTmpL2[i] = pRec[i*iStride+iShift];
      //pRec += iStride;
    }
    //pRec -= (iStride*(iCuHeightTmp+1));

    pTmpL = this.m_pTmpL1; 
	pTmpU = this.m_pTmpU1;//[uiLPelX:]; 
  }

  if iYCbCr==0 {
  	pClipTbl = this.m_pClipTable;
  	pOffsetBo = this.m_iOffsetBo;
  }else{
  	pClipTbl = this.m_pChromaClipTable;
  	pOffsetBo = this.m_iChromaOffsetBo;
  }

  switch iSaoType{
  case SAO_EO_0: // dir: -
      iStartX = int(B2U(uiLPelX == 0));
      if uiRPelX == uint(iPicWidthTmp) {
      	iEndX   = iLcuWidth-1;
      }else{
      	iEndX   = iLcuWidth;
      }
      
      for y=0; y<iLcuHeight; y++ {
        iSignLeft = this.xSign(int(pRec[y*iStride+iStartX] - pTmpL[y]));
        for x=iStartX; x< iEndX; x++ {
          iSignRight =  this.xSign(int(pRec[y*iStride+x] - pRec[y*iStride+x+1])); 
          uiEdgeType =  uint(iSignRight + iSignLeft + 2);
          iSignLeft  = -iSignRight;

          pRec[y*iStride+x] = pClipTbl[int(pRec[y*iStride+x]) + this.m_iOffsetEo[uiEdgeType]];
        }
        //pRec += iStride;
      }
  case SAO_EO_1: // dir: |
      iStartY = int(B2U(uiTPelY == 0));
      if uiBPelY == uint(iPicHeightTmp) {
      	iEndY   = iLcuHeight-1;
      }else{
      	iEndY   = iLcuHeight;
      }
      if uiTPelY == 0 {
        pRec = pRec[ iStride:];
      }
      for x=0; x< iLcuWidth; x++ {
        this.m_iUpBuff1[1+x] = this.xSign(int(pRec[x] - pTmpU[int(uiLPelX)+x]));
      }
      for y=iStartY; y<iEndY; y++ {
        for x=0; x<iLcuWidth; x++ {
          iSignDown  = this.xSign(int(pRec[y*iStride+x] - pRec[y*iStride+x+iStride])); 
          uiEdgeType = uint(iSignDown + this.m_iUpBuff1[1+x] + 2);
          this.m_iUpBuff1[1+x]= -iSignDown;

          pRec[y*iStride+x] = pClipTbl[int(pRec[y*iStride+x]) + this.m_iOffsetEo[uiEdgeType]];
        }
        //pRec += iStride;
      }
  case SAO_EO_2: // dir: 135
      iStartX = int(B2U(uiLPelX == 0));
      if uiRPelX == uint(iPicWidthTmp) {
      	iEndX   = iLcuWidth-1;
	  }else{
	  	iEndX   = iLcuWidth;
	  }
	  
      iStartY = int(B2U(uiTPelY == 0));
      if uiBPelY == uint(iPicHeightTmp) {
      	iEndY   = iLcuHeight-1;
      }else{
      	iEndY   = iLcuHeight;
      }

      if uiTPelY == 0 {
        pRec = pRec[iStride:];
      }

      for x=iStartX; x<iEndX; x++ {
        this.m_iUpBuff1[1+x] = this.xSign(int(pRec[x] - pTmpU[int(uiLPelX)+x-1]));
      }
      for y=iStartY; y<iEndY; y++ {
        iSignDown2 = this.xSign(int(pRec[y*iStride+iStride+iStartX] - pTmpL[y]));
        for x=iStartX; x<iEndX; x++ {
          iSignDown1      =  this.xSign(int(pRec[y*iStride+x] - pRec[y*iStride+x+iStride+1])) ;
          uiEdgeType      =  uint(iSignDown1 + this.m_iUpBuff1[1+x] + 2);
          this.m_iUpBufft[1+x+1] = -iSignDown1; 
          pRec[y*iStride+x] = pClipTbl[int(pRec[y*iStride+x]) + this.m_iOffsetEo[uiEdgeType]];
        }
        this.m_iUpBufft[1+iStartX] = iSignDown2;

        this.m_ipSwap     = this.m_iUpBuff1;
        this.m_iUpBuff1 = this.m_iUpBufft;
        this.m_iUpBufft = this.m_ipSwap;

        //pRec += iStride;
      }
  case SAO_EO_3: // dir: 45
      iStartX = int(B2U(uiLPelX == 0));
      if uiRPelX == uint(iPicWidthTmp) {
      	iEndX   = iLcuWidth-1 ;
      }else{
      	iEndX   = iLcuWidth;
      }

      iStartY = int(B2U(uiTPelY == 0));
      if uiBPelY == uint(iPicHeightTmp) {
      	iEndY   = iLcuHeight-1;
      }else{
      	iEndY   = iLcuHeight;
      }

      if iStartY == 1 {
        pRec = pRec[iStride:];
      }

      for x=iStartX-1; x<iEndX; x++ {
        this.m_iUpBuff1[1+x] = this.xSign(int(pRec[x] - pTmpU[int(uiLPelX)+x+1]));
      }
      for y=iStartY; y<iEndY; y++ {
        x=iStartX;
        iSignDown1      =  this.xSign(int(pRec[y*iStride+x] - pTmpL[y+1])) ;
        uiEdgeType      =  uint(iSignDown1 + this.m_iUpBuff1[1+x]) + 2;
        this.m_iUpBuff1[1+x-1] = -iSignDown1; 
        pRec[y*iStride+x] = pClipTbl[int(pRec[y*iStride+x]) + this.m_iOffsetEo[uiEdgeType]];
        for x=iStartX+1; x<iEndX; x++ {
          iSignDown1      =  this.xSign(int(pRec[y*iStride+x] - pRec[y*iStride+x+iStride-1])) ;
          uiEdgeType      =  uint(iSignDown1 + this.m_iUpBuff1[1+x] + 2);
          this.m_iUpBuff1[1+x-1] = -iSignDown1; 
          pRec[y*iStride+x] = pClipTbl[int(pRec[y*iStride+x]) + this.m_iOffsetEo[uiEdgeType]];
        }
        this.m_iUpBuff1[1+iEndX-1] = this.xSign(int(pRec[y*iStride+iEndX-1 + iStride] - pRec[y*iStride+iEndX]));

        //pRec += iStride;
      }   
  case SAO_BO:
      for y=0; y<iLcuHeight; y++ {
        for x=0; x<iLcuWidth; x++ {
          pRec[y*iStride+x] = Pel(pOffsetBo[pRec[y*iStride+x]]);
        }
        //pRec += iStride;
      }
  default: //break;
  }
//   if (iSaoType!=SAO_BO_0 || iSaoType!=SAO_BO_1)
  {
    pTmpLSwap = this.m_pTmpL1;
    this.m_pTmpL1  = this.m_pTmpL2;
    this.m_pTmpL2  = pTmpLSwap;
  }
}
func (this *TComSampleAdaptiveOffset) CreatePicSaoInfo(pcPic *TComPic,  numSlicesInPic int){
  this.m_pcPic   = pcPic;
  this.m_uiNumSlicesInPic = uint(numSlicesInPic);
  this.m_iSGDepth = 0;
  this.m_bUseNIF = ( pcPic.GetIndependentSliceBoundaryForNDBFilter() || pcPic.GetIndependentTileBoundaryForNDBFilter() );
  if this.m_bUseNIF {
    this.m_pcYuvTmp = pcPic.GetYuvPicBufferForIndependentBoundaryProcessing();
  }
}
func (this *TComSampleAdaptiveOffset) DestroyPicSaoInfo(){
	//do nothing
}
func (this *TComSampleAdaptiveOffset) ProcessSaoBlock(pDec []Pel, pRest []Pel,  stride,  saoType int,  width,  height int, pbBorderAvail []bool,  iYCbCr int){ 
  //variables
  var startX, startY, endX, endY, x, y int;
  var signLeft,signRight,signDown,signDown1 int;
  var edgeType uint;
  var pClipTbl []Pel;
  var pOffsetBo []int;
  
  if iYCbCr==0 {
  	pClipTbl = this.m_pClipTable;
    pOffsetBo = this.m_iOffsetBo;
  }else{
    pClipTbl = this.m_pChromaClipTable;
    pOffsetBo = this.m_iChromaOffsetBo;
  }
  
  
  switch saoType {
  case SAO_EO_0: // dir: -
  	  if pbBorderAvail[SGU_L]{
      	startX =  0;
      }else{
      	startX =  1;
      }
      if pbBorderAvail[SGU_R] {
      	endX   = width;
      }else{	
      	endX   = (width -1);
      }
      for y=0; y< height; y++{
        signLeft = this.xSign(int(pDec[y*stride+startX] - pDec[y*stride+startX-1]));
        for x=startX; x< endX; x++{
          signRight =  this.xSign(int(pDec[y*stride+x] - pDec[y*stride+x+1])); 
          edgeType =  uint(signRight + signLeft + 2);
          signLeft  = -signRight;

          pRest[y*stride+x] = pClipTbl[int(pDec[y*stride+x]) + this.m_iOffsetEo[edgeType]];
        }
        //pDec  += stride;
        //pRest += stride;
      }
  case SAO_EO_1: // dir: |
  	  if pbBorderAvail[SGU_T]{
      	startY = 0;
      }else{
      	startY = 1;
      }
      if pbBorderAvail[SGU_B]{
      	endY   = height;
      }else{
      	endY   = height-1;
      }
      for x=0; x< width; x++ {
        this.m_iUpBuff1[1+x] = this.xSign(int(pDec[x+stride] - pDec[x]));
      }
      if !pbBorderAvail[SGU_T]{
        pDec  = pDec[stride:];
        pRest = pRest[stride:];
      }
      for y=startY; y<endY; y++ {
        for x=0; x< width; x++ {
          signDown  = this.xSign(int(pDec[y*stride+x] - pDec[y*stride+x+stride])); 
          edgeType = uint(signDown + this.m_iUpBuff1[1+x] + 2);
          this.m_iUpBuff1[1+x]= -signDown;

          pRest[y*stride+x] = pClipTbl[int(pDec[y*stride+x]) + this.m_iOffsetEo[edgeType]];
        }
        //pDec  += stride;
        //pRest += stride;
      }
  case SAO_EO_2: // dir: 135
      posShift := stride + 1;

	  if pbBorderAvail[SGU_L] {
      	startX = 0;
      }else{
      	startX = 1;
      }
      if pbBorderAvail[SGU_R] {
      	endX   = width;
      }else{
      	endX   = (width-1);
      }

      //prepare 2nd line upper sign
      pDec2 := pDec[ stride:];
      for x=startX; x< endX+1; x++ {
        this.m_iUpBuff1[1+x] = this.xSign(int(pDec2[x] - pDec2[x- posShift]));
      }

      //1st line
      //pDec -= stride;
      if pbBorderAvail[SGU_TL] {
        x= 0;
        edgeType =  uint(this.xSign(int(pDec[x] - pDec[x- posShift])) - this.m_iUpBuff1[1+x+1] + 2);
        pRest[x] = pClipTbl[int(pDec[x]) + this.m_iOffsetEo[edgeType]];
      }
      if pbBorderAvail[SGU_T] {
        for x= 1; x< endX; x++ {
          edgeType = uint(this.xSign(int(pDec[x] - pDec[x- posShift])) - this.m_iUpBuff1[1+x+1] + 2);
          pRest[x] = pClipTbl[int(pDec[x]) + this.m_iOffsetEo[edgeType]];
        }
      }
      pDec   = pDec[ stride:];
      pRest  = pRest[stride:];


      //middle lines
      for y= 1; y< height-1; y++ {
        for x=startX; x<endX; x++ {
          signDown1     =  this.xSign(int(pDec[y*stride+x] - pDec[y*stride+x+ posShift])) ;
          edgeType      =  uint(signDown1 + this.m_iUpBuff1[1+x] + 2);
          pRest[y*stride+x] = pClipTbl[int(pDec[y*stride+x]) + this.m_iOffsetEo[edgeType]];

          this.m_iUpBufft[1+x+1] = -signDown1; 
        }
        this.m_iUpBufft[1+startX] = this.xSign(int(pDec[y*stride+stride+startX] - pDec[y*stride+startX-1]));

        this.m_ipSwap   = this.m_iUpBuff1;
        this.m_iUpBuff1 = this.m_iUpBufft;
        this.m_iUpBufft = this.m_ipSwap;

        //pDec  += stride;
        //pRest += stride;
      }

      //last line
      if pbBorderAvail[SGU_B] {
        for x= startX; x< width-1; x++ {
          edgeType =  uint(this.xSign(int(pDec[x] - pDec[x+ posShift])) + this.m_iUpBuff1[1+x] + 2);
          pRest[x] = pClipTbl[int(pDec[x]) + this.m_iOffsetEo[edgeType]];
        }
      }
      if pbBorderAvail[SGU_BR] {
        x= width -1;
        edgeType =  uint(this.xSign(int(pDec[x] - pDec[x+ posShift])) + this.m_iUpBuff1[1+x] + 2);
        pRest[x] = pClipTbl[int(pDec[x]) + this.m_iOffsetEo[edgeType]];
      }
  case SAO_EO_3: // dir: 45
      posShift := stride - 1;
      if pbBorderAvail[SGU_L]{
      	startX = 0;
      }else{
      	startX = 1;
      }
      if pbBorderAvail[SGU_R]{
      	endX   = width;
      }else{
      	endX   = (width -1);
      }

      //prepare 2nd line upper sign
      pDec2 := pDec[stride:];
      for x=startX-1; x< endX; x++{
        this.m_iUpBuff1[1+x] = this.xSign(int(pDec2[x] - pDec[stride+x- posShift]));
      }

      //first line
      //pDec -= stride;
      if pbBorderAvail[SGU_T] {
        for x= startX; x< width -1; x++ {
          edgeType = uint(this.xSign(int(pDec[x] - pDec[x- posShift])) -this.m_iUpBuff1[1+x-1] + 2);
          pRest[x] = pClipTbl[int(pDec[x]) + this.m_iOffsetEo[edgeType]];
        }
      }
      if pbBorderAvail[SGU_TR] {
        x= width-1;
        edgeType = uint(this.xSign(int(pDec[x] - pDec[x- posShift])) -this.m_iUpBuff1[1+x-1] + 2);
        pRest[x] = pClipTbl[int(pDec[x]) + this.m_iOffsetEo[edgeType]];
      }
      pDec  = pDec[stride:];
      pRest = pRest[stride:];

      //middle lines
      for y= 1; y< height-1; y++ {
        for x= startX; x< endX; x++ {
          signDown1     =  this.xSign(int(pDec[y*stride+x] - pDec[y*stride+x+ posShift])) ;
          edgeType      =  uint(signDown1 + this.m_iUpBuff1[1+x] + 2);

          pRest[y*stride+x] = pClipTbl[int(pDec[y*stride+x]) + this.m_iOffsetEo[edgeType]];
          this.m_iUpBuff1[1+x-1] = -signDown1; 
        }
        this.m_iUpBuff1[1+endX-1] = this.xSign(int(pDec[y*stride+endX-1 + stride] - pDec[y*stride+endX]));

        //pDec  += stride;
        //pRest += stride;
      }

      //last line
      if pbBorderAvail[SGU_BL] {
        x= 0;
        edgeType = uint(this.xSign(int(pDec[x] - pDec[x+ posShift])) + this.m_iUpBuff1[1+x] + 2);
        pRest[x] = pClipTbl[int(pDec[x]) + this.m_iOffsetEo[edgeType]];

      }
      if pbBorderAvail[SGU_B] {
        for x= 1; x< endX; x++ {
          edgeType = uint(this.xSign(int(pDec[x] - pDec[x+ posShift])) + this.m_iUpBuff1[1+x] + 2);
          pRest[x] = pClipTbl[int(pDec[x]) + this.m_iOffsetEo[edgeType]];
        }
      }  
  case SAO_BO:
      for y=0; y< height; y++ {
        for x=0; x< width; x++ {
          pRest[y*stride+x] = Pel(pOffsetBo[pDec[y*stride+x]]);
        }
        //pRest += stride;
        //pDec  += stride;
      }
  default: //break;
  } 
}

func (this *TComSampleAdaptiveOffset) ResetLcuPart(saoLcuParam []SaoLcuParam){
  var i,j int;
  for i=0;i<this.m_iNumCuInWidth*this.m_iNumCuInHeight;i++ {
    saoLcuParam[i].MergeUpFlag  =  true;
    saoLcuParam[i].MergeLeftFlag =  false;
    saoLcuParam[i].PartIdx   =  0;
    saoLcuParam[i].TypeIdx      = -1;
    for j=0;j<MAX_NUM_SAO_OFFSETS;j++ {
      saoLcuParam[i].Offset[j] = 0;
    }
    saoLcuParam[i].SubTypeIdx = 0;
  }
}
func (this *TComSampleAdaptiveOffset) ConvertQT2SaoUnit(saoParam *SAOParam,  partIdx int, yCbCr int){
  saoPart := &(saoParam.SaoPart[yCbCr][partIdx]);
  if !saoPart.bSplit {
    this.ConvertOnePart2SaoUnit(saoParam, partIdx, yCbCr);
    return;
  }

  if saoPart.PartLevel < int(this.m_uiMaxSplitLevel) {
    this.ConvertQT2SaoUnit(saoParam, saoPart.DownPartsIdx[0], yCbCr);
    this.ConvertQT2SaoUnit(saoParam, saoPart.DownPartsIdx[1], yCbCr);
    this.ConvertQT2SaoUnit(saoParam, saoPart.DownPartsIdx[2], yCbCr);
    this.ConvertQT2SaoUnit(saoParam, saoPart.DownPartsIdx[3], yCbCr);
  }
}
func (this *TComSampleAdaptiveOffset) ConvertOnePart2SaoUnit(saoParam *SAOParam,  partIdx int,  yCbCr int){
  var j, idxX, idxY, addr int;
  frameWidthInCU := int(this.m_pcPic.GetFrameWidthInCU());
  saoQTPart := saoParam.SaoPart[yCbCr];
  saoLcuParam := saoParam.SaoLcuParam[yCbCr];

  for idxY = saoQTPart[partIdx].StartCUY; idxY<= saoQTPart[partIdx].EndCUY; idxY++ {
    for idxX = saoQTPart[partIdx].StartCUX; idxX<= saoQTPart[partIdx].EndCUX; idxX++ {
      addr = idxY * frameWidthInCU + idxX;
      saoLcuParam[addr].PartIdxTmp = partIdx; 
      saoLcuParam[addr].TypeIdx    = saoQTPart[partIdx].iBestType;
      saoLcuParam[addr].SubTypeIdx = saoQTPart[partIdx].subTypeIdx;
      if saoLcuParam[addr].TypeIdx!=-1 {
        saoLcuParam[addr].Length    = saoQTPart[partIdx].iLength;
        for j=0;j<MAX_NUM_SAO_OFFSETS;j++ {
          saoLcuParam[addr].Offset[j] = saoQTPart[partIdx].iOffset[j];
        }
      }else{
        saoLcuParam[addr].Length    = 0;
        saoLcuParam[addr].SubTypeIdx = saoQTPart[partIdx].subTypeIdx;
        for j=0;j<MAX_NUM_SAO_OFFSETS;j++ {
          saoLcuParam[addr].Offset[j] = 0;
        }
      }
    }
  }
}
func (this *TComSampleAdaptiveOffset) ProcessSaoUnitAll(saoLcuParam []SaoLcuParam, oneUnitFlag bool,  yCbCr int){
  var pRec []Pel;
  var picWidthTmp int;

  if yCbCr == 0 {
    pRec        = this.m_pcPic.GetPicYuvRec().GetLumaAddr();
    picWidthTmp = this.m_iPicWidth;
  }else if yCbCr == 1 {
    pRec        = this.m_pcPic.GetPicYuvRec().GetCbAddr();
    picWidthTmp = this.m_iPicWidth>>1;
  }else{
    pRec        = this.m_pcPic.GetPicYuvRec().GetCrAddr();
    picWidthTmp = this.m_iPicWidth>>1;
  }

  for i:=0; i<picWidthTmp; i++{
  	this.m_pTmpU1[i] = pRec[i];//, sizeof(Pel)*picWidthTmp);
  }

  var  i int;
  var edgeType uint;
  var ppLumaTable []Pel;// = NULL;
  var pClipTable []Pel;//= NULL;
  var pOffsetBo []int;//= NULL;
  var  typeIdx int;

  var offset	[LUMA_GROUP_NUM+1]int;
  var idxX,idxY,addr int;
  frameWidthInCU := int(this.m_pcPic.GetFrameWidthInCU());
  frameHeightInCU := int(this.m_pcPic.GetFrameHeightInCU());
  var stride int;
  var tmpUSwap []Pel;
  isChroma := B2U(yCbCr != 0);
  var mergeLeftFlag bool;
  var saoBitIncrease int;
  if yCbCr == 0 {
  	saoBitIncrease = int(this.m_uiSaoBitIncreaseY);
  	pOffsetBo = this.m_iOffsetBo;
  }else{
  	saoBitIncrease = int(this.m_uiSaoBitIncreaseC);
  	pOffsetBo = this.m_iChromaOffsetBo;
  }

  offset[0] = 0;
  for idxY = 0; idxY< frameHeightInCU; idxY++ { 
    addr = idxY * frameWidthInCU;
    if yCbCr == 0 {
      pRec  = this.m_pcPic.GetPicYuvRec().GetLumaAddr1(addr);
      stride = this.m_pcPic.GetStride();
      picWidthTmp = this.m_iPicWidth;
    }else if yCbCr == 1 {
      pRec  = this.m_pcPic.GetPicYuvRec().GetCbAddr1(addr);
      stride = this.m_pcPic.GetCStride();
      picWidthTmp = this.m_iPicWidth>>1;
    }else{
      pRec  = this.m_pcPic.GetPicYuvRec().GetCrAddr1(addr);
      stride = this.m_pcPic.GetCStride();
      picWidthTmp = this.m_iPicWidth>>1;
    }

    for i=0;i<int(this.m_uiMaxCUHeight>>isChroma)+1;i++ {
      this.m_pTmpL1[i] = pRec[i*stride+0];
      //pRec+=stride;
    }
    pRec=pRec[(int(this.m_uiMaxCUHeight>>isChroma))*stride-(stride<<1):];

	for i=0; i<picWidthTmp; i++{
    	this.m_pTmpU2[i] = pRec[i];//, sizeof(Pel)*picWidthTmp);
    }

    for idxX = 0; idxX < frameWidthInCU; idxX++ {
      addr = idxY * frameWidthInCU + idxX;

      if oneUnitFlag{
        typeIdx = saoLcuParam[0].TypeIdx;
        mergeLeftFlag = (addr != 0);
      }else{
        typeIdx = saoLcuParam[addr].TypeIdx;
        mergeLeftFlag = saoLcuParam[addr].MergeLeftFlag;
      }
      if typeIdx>=0 {
        if !mergeLeftFlag {

          if typeIdx == SAO_BO{
            for i=0; i<SAO_MAX_BO_CLASSES+1;i++ {
              offset[i] = 0;
            }
            for i=0; i<saoLcuParam[addr].Length; i++ {
              offset[ (saoLcuParam[addr].SubTypeIdx +i)%SAO_MAX_BO_CLASSES  +1] = saoLcuParam[addr].Offset[i] << uint(saoBitIncrease);
            }
			
			var bitDepth int;
			if yCbCr==0 {
	            ppLumaTable = this.m_lumaTableBo;
	            pClipTable = this.m_pClipTable;
	            bitDepth = G_bitDepthY;
            }else{
            	ppLumaTable = this.m_chromaTableBo;
	            pClipTable = this.m_pChromaClipTable;
	            bitDepth = G_bitDepthC;
            }

            for i=0;i<(1<<uint(bitDepth));i++ {
              pOffsetBo[i] = int(pClipTable[i + offset[ppLumaTable[i]]]);
            }

          }
          if typeIdx == SAO_EO_0 || typeIdx == SAO_EO_1 || typeIdx == SAO_EO_2 || typeIdx == SAO_EO_3 {
            for i=0;i<saoLcuParam[addr].Length;i++ {
              offset[i+1] = saoLcuParam[addr].Offset[i] << uint(saoBitIncrease);
            }
            for edgeType=0;edgeType<6;edgeType++ {
              this.m_iOffsetEo[edgeType]= offset[m_auiEoTable[edgeType]];
            }
          }
        }
        this.ProcessSaoCu(addr, typeIdx, yCbCr);
      }else{
        if idxX != (frameWidthInCU-1) {
          if yCbCr == 0 {
            pRec  = this.m_pcPic.GetPicYuvRec().GetLumaAddr1(addr);
            stride = this.m_pcPic.GetStride();
          }else if yCbCr == 1 {
            pRec  = this.m_pcPic.GetPicYuvRec().GetCbAddr1(addr);
            stride = this.m_pcPic.GetCStride();
          }else{
            pRec  = this.m_pcPic.GetPicYuvRec().GetCrAddr1(addr);
            stride = this.m_pcPic.GetCStride();
          }
          widthShift := this.m_uiMaxCUWidth>>isChroma;
          for i=0;i<int(this.m_uiMaxCUHeight>>isChroma)+1;i++ {
            this.m_pTmpL1[i] = pRec[i*stride+int(widthShift)-1];
            //pRec+=stride;
          }
        }
      }
    }
    tmpUSwap = this.m_pTmpU1;
    this.m_pTmpU1 = this.m_pTmpU2;
    this.m_pTmpU2 = tmpUSwap;
  }
}
func (this *TComSampleAdaptiveOffset) SetSaoLcuBoundary ( bVal bool)  {
	this.m_saoLcuBoundary = bVal;
}
func (this *TComSampleAdaptiveOffset) GetSaoLcuBoundary ()    bool       {
	return this.m_saoLcuBoundary;
}
func (this *TComSampleAdaptiveOffset) SetSaoLcuBasedOptimization ( bVal bool)  {
	this.m_saoLcuBasedOptimization = bVal;
}
func (this *TComSampleAdaptiveOffset) GetSaoLcuBasedOptimization ()  bool         {
	return this.m_saoLcuBasedOptimization;
}
func (this *TComSampleAdaptiveOffset) ResetSaoUnit(saoUnit *SaoLcuParam){
  saoUnit.PartIdx       = 0;
  saoUnit.PartIdxTmp    = 0;
  saoUnit.MergeLeftFlag = false;
  saoUnit.MergeUpFlag   = false;
  saoUnit.TypeIdx       = -1;
  saoUnit.Length        = 0;
  saoUnit.SubTypeIdx    = 0;

  for i:=0;i<4;i++ {
    saoUnit.Offset[i] = 0;
  }
}
func (this *TComSampleAdaptiveOffset) CopySaoUnit(saoUnitDst *SaoLcuParam, saoUnitSrc *SaoLcuParam){
  saoUnitDst.MergeLeftFlag = saoUnitSrc.MergeLeftFlag;
  saoUnitDst.MergeUpFlag   = saoUnitSrc.MergeUpFlag;
  saoUnitDst.TypeIdx       = saoUnitSrc.TypeIdx;
  saoUnitDst.Length        = saoUnitSrc.Length;

  saoUnitDst.SubTypeIdx  = saoUnitSrc.SubTypeIdx;
  for i:=0;i<4;i++ {
    saoUnitDst.Offset[i] = saoUnitSrc.Offset[i];
  }
}
func (this *TComSampleAdaptiveOffset) PCMLFDisableProcess    ( pcPic *TComPic ){                       ///< interface function for ALF process
  this.xPCMRestoration(pcPic);
}

func (this *TComSampleAdaptiveOffset) xPCMRestoration        (pcPic *TComPic){
  bPCMFilter := pcPic.GetSlice(0).GetSPS().GetUsePCM() && pcPic.GetSlice(0).GetSPS().GetPCMFilterDisableFlag();

  if bPCMFilter || pcPic.GetSlice(0).GetPPS().GetTransquantBypassEnableFlag() {
    for uiCUAddr := uint(0); uiCUAddr < pcPic.GetNumCUsInFrame() ; uiCUAddr++ {
      pcCU := pcPic.GetCU(uiCUAddr);

      this.xPCMCURestoration(pcCU, 0, 0); 
    } 
  }
}
func (this *TComSampleAdaptiveOffset) xPCMCURestoration      (pcCU *TComDataCU,  uiAbsZorderIdx,  uiDepth uint){
  pcPic     := pcCU.GetPic();
  uiCurNumParts := pcPic.GetNumPartInCU() >> (uiDepth<<1);
  uiQNumParts   := uiCurNumParts>>2;

  // go to sub-CU
  if uint(pcCU.GetDepth1(uiAbsZorderIdx)) > uiDepth {
    for uiPartIdx := uint(0); uiPartIdx < 4; uiPartIdx++ {
      uiLPelX   := pcCU.GetCUPelX() + G_auiRasterToPelX[ G_auiZscanToRaster[uiAbsZorderIdx] ];
      uiTPelY   := pcCU.GetCUPelY() + G_auiRasterToPelY[ G_auiZscanToRaster[uiAbsZorderIdx] ];
      if ( uiLPelX < pcCU.GetSlice().GetSPS().GetPicWidthInLumaSamples() ) && ( uiTPelY < pcCU.GetSlice().GetSPS().GetPicHeightInLumaSamples() ) {
        this.xPCMCURestoration( pcCU, uiAbsZorderIdx, uiDepth+1 );
      }
      uiAbsZorderIdx+=uiQNumParts;
    }
    return;
  }

  // restore PCM samples
  if (pcCU.GetIPCMFlag1(uiAbsZorderIdx)&& pcPic.GetSlice(0).GetSPS().GetPCMFilterDisableFlag()) || pcCU.IsLosslessCoded( uiAbsZorderIdx) {
    this.xPCMSampleRestoration (pcCU, uiAbsZorderIdx, uiDepth, TEXT_LUMA    );
    this.xPCMSampleRestoration (pcCU, uiAbsZorderIdx, uiDepth, TEXT_CHROMA_U);
    this.xPCMSampleRestoration (pcCU, uiAbsZorderIdx, uiDepth, TEXT_CHROMA_V);
  }
}
func (this *TComSampleAdaptiveOffset) xPCMSampleRestoration  (pcCU *TComDataCU,  uiAbsZorderIdx,  uiDepth uint,  ttText TextType){
  pcPicYuvRec := pcCU.GetPic().GetPicYuvRec();
  var piSrc []Pel;
  var piPcm []Pel;
  var uiStride,uiWidth,uiHeight,uiPcmLeftShiftBit,uiX, uiY	uint;
  uiMinCoeffSize := pcCU.GetPic().GetMinCUWidth()*pcCU.GetPic().GetMinCUHeight();
  uiLumaOffset   := uiMinCoeffSize*uiAbsZorderIdx;
  uiChromaOffset := uiLumaOffset>>2;

  if ttText == TEXT_LUMA {
    piSrc = pcPicYuvRec.GetLumaAddr2( int(pcCU.GetAddr()), int(uiAbsZorderIdx) );
    piPcm = pcCU.GetPCMSampleY() [ uiLumaOffset:];
    uiStride  = uint(pcPicYuvRec.GetStride());
    uiWidth  = (G_uiMaxCUWidth >> uiDepth);
    uiHeight = (G_uiMaxCUHeight >> uiDepth);
    if pcCU.IsLosslessCoded(uiAbsZorderIdx) && !pcCU.GetIPCMFlag1(uiAbsZorderIdx) {
      uiPcmLeftShiftBit = 0;
    }else{
      uiPcmLeftShiftBit = uint(G_bitDepthY) - pcCU.GetSlice().GetSPS().GetPCMBitDepthLuma();
    }
  }else{
    if ttText == TEXT_CHROMA_U {
      piSrc = pcPicYuvRec.GetCbAddr2( int(pcCU.GetAddr()), int(uiAbsZorderIdx) );
      piPcm = pcCU.GetPCMSampleCb() [ uiChromaOffset:];
    }else{
      piSrc = pcPicYuvRec.GetCrAddr2( int(pcCU.GetAddr()), int(uiAbsZorderIdx) );
      piPcm = pcCU.GetPCMSampleCr() [ uiChromaOffset:];
    }

    uiStride = uint(pcPicYuvRec.GetCStride());
    uiWidth  = ((G_uiMaxCUWidth >> uiDepth)/2);
    uiHeight = ((G_uiMaxCUWidth >> uiDepth)/2);
    if pcCU.IsLosslessCoded(uiAbsZorderIdx) && !pcCU.GetIPCMFlag1(uiAbsZorderIdx) {
      uiPcmLeftShiftBit = 0;
    }else{
      uiPcmLeftShiftBit = uint(G_bitDepthC) - pcCU.GetSlice().GetSPS().GetPCMBitDepthChroma();
    }
  }

  for uiY = 0; uiY < uiHeight; uiY++ {
    for uiX = 0; uiX < uiWidth; uiX++ {
      piSrc[uiY*uiStride+uiX] = (piPcm[uiY*uiWidth+uiX] << uiPcmLeftShiftBit);
    }
    //piPcm += uiWidth;
    //piSrc += uiStride;
  }
}