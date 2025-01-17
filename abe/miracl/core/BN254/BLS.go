/*
   Copyright (C) 2019 MIRACL UK Ltd.

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as
    published by the Free Software Foundation, either version 3 of the
    License, or (at your option) any later version.


    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

     https://www.gnu.org/licenses/agpl-3.0.en.html

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

   You can be released from the requirements of the license by purchasing
   a commercial license. Buying such a license is mandatory as soon as you
   develop commercial activities involving the MIRACL Core Crypto SDK
   without disclosing the source code of your own applications, or shipping
   the MIRACL Core Crypto SDK with a closed source product.
*/

/* Boneh-Lynn-Shacham  API Functions */

/* Loosely (for now) following https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-bls-signature-00 */

// Minimal-signature-size variant

package BN254

//import "github.com/miracl/core/go/core"
import "github.com/marcellop71/mosaic/abe/miracl/core"

import "fmt"

const BFS int = int(MODBYTES)
const BGS int = int(MODBYTES)
const BLS_OK int = 0
const BLS_FAIL int = -1

var G2_TAB []*FP4

func ceil(a int,b int) int {
    return (((a)-1)/(b)+1)
}

/* output u \in F_p */
func hash_to_field(hash int,hlen int ,DST []byte,M []byte,ctr int) []*FP {
	q := NewBIGints(Modulus)
	L := ceil(q.nbits()+AESKEY*8,8)
	var u []*FP
	var fd =make([]byte,L)
	OKM:=core.XMD_Expand(hash,hlen,L*ctr,DST,M)

	for i:=0;i<ctr;i++ {
		for j:=0;j<L;j++ {
			fd[j]=OKM[i*L+j];
		}
		u = append(u,NewFPbig(DBIG_fromBytes(fd).Mod(q)))
	}
	return u
}

/* output u \in F_p
func hash_to_base(hash int,hlen int ,DST []byte,M []byte,ctr int) *BIG {
	q := NewBIGints(Modulus)
	L := ceil(q.nbits()+AESKEY*8,8)
	INFO := []byte("H2C")
	INFO = append(INFO,byte(ctr))

	PRK:=core.HKDF_Extract(hash,hlen,DST,M)
	OKM:=core.HKDF_Expand(hash,hlen,L,PRK,INFO)

	dx:= DBIG_fromBytes(OKM[:])
	u:= dx.mod(q)
	return u
} */


/* hash a message to an ECP point, using SHA2, random oracle method */
func bls_hash_to_point(M []byte) *ECP {
	DST := []byte("BLS_SIG_ZZZG1_XMD:SHA256-SSWU-RO-_NUL_")
	u := hash_to_field(core.MC_SHA2,HASH_TYPE,DST,M,2)

	P:=ECP_map2point(u[0])
	P1 := ECP_map2point(u[1]);
	P.Add(P1)
	P.Cfp()
	P.Affine()
	return P
}

func Init() int {
	G := ECP2_generator()
	if G.Is_infinity() {
		return BLS_FAIL
	}
	G2_TAB = precomp(G)
	return BLS_OK
}

/* generate key pair, private key S, public key W */
func KeyPairGenerate(IKM []byte, S []byte, W []byte) int {
	r := NewBIGints(CURVE_Order)
	L := ceil(3*ceil(r.nbits(),8),2)
	G := ECP2_generator()
	if G.Is_infinity() {
		return BLS_FAIL
	}
	SALT := []byte("BLS-SIG-KEYGEN-SALT-")
	INFO := []byte("")
	PRK := core.HKDF_Extract(core.MC_SHA2,HASH_TYPE,SALT,IKM)
	OKM := core.HKDF_Expand(core.MC_SHA2,HASH_TYPE,L,PRK,INFO)

	dx:= DBIG_fromBytes(OKM[:])
	s:= dx.Mod(r)

	s.ToBytes(S)
	G = G2mul(G, s)
	G.ToBytes(W,true)
	return BLS_OK
}

/* Sign message M using private key S to produce signature SIG */

func Core_Sign(SIG []byte, M []byte, S []byte) int {
	D := bls_hash_to_point(M)
	s := FromBytes(S)
	D = G1mul(D, s)
	D.ToBytes(SIG, true)
	return BLS_OK
}

/* Verify signature given message m, the signature SIG, and the public key W */

func Core_Verify(SIG []byte, M []byte, W []byte) int {
	HM := bls_hash_to_point(M)

	D := ECP_fromBytes(SIG)
	if !G1member(D) {return BLS_FAIL}
	D.Neg()

	PK := ECP2_fromBytes(W)

	// Use new multi-pairing mechanism

	r := Initmp()
	Another_pc(r, G2_TAB, D)
	Another(r, PK, HM)
	v := Miller(r)

	//.. or alternatively
	//	G := ECP2_generator()
	//	if G.Is_infinity() {return BLS_FAIL}
	//	v := Ate2(G, D, PK, HM)

	v = Fexp(v)

	if v.Isunity() {
		return BLS_OK
	} else {
		return BLS_FAIL
	}
}

// ------------------------------------------
// added by Marcello Paris
// for github.com/marcellop71/mosaic
//
// BEGIN
// ------------------------------------------

func Hash_to_point_G1(M []byte) *ECP {
  return bls_hash_to_point(M)
}

/* hash a message to an ECP point, using SHA2, random oracle method */
func Hash_to_point_G2(M []byte) *ECP2 {
	DST := []byte("BLS_SIG_ZZZG1_XMD:SHA256-SSWU-RO-_NUL_")
	fmt.Printf("\nhash2point\n")
	u := hash_to_field(core.MC_SHA2,HASH_TYPE,DST,M,4)
	
	P:=ECP2_map2point(NewFP2fps(u[0], u[1]))
	P1 := ECP2_map2point(NewFP2fps(u[2], u[3]))
	P.Add(P1)
	P.Cfp()
	P.Affine()
	return P
}

// ------------------------------------------
// END
// ------------------------------------------
